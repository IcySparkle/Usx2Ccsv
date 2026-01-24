use csv::Writer;
use glob::glob;
use quick_xml::events::Event;
use quick_xml::Reader;
use regex::{Captures, Regex};
use serde::Serialize;
use std::collections::HashMap;
use std::env;
use std::fs;
use std::io;
use std::path::{Path, PathBuf};

#[derive(Clone, PartialEq)]
enum NodeType {
    Element,
    Text,
}

#[derive(Clone)]
struct Node {
    node_type: NodeType,
    name: String,
    attrs: HashMap<String, String>,
    children: Vec<Node>,
    text: String,
}

struct Row {
    book: String,
    chapter: String,
    verse: String,
    text_plain: String,
    text_styled: String,
    footnotes: String,
    crossrefs: String,
    subtitle: String,
}

#[derive(Serialize)]
struct FileResult {
    input: String,
    output: String,
    format: String,
    rows: usize,
}

#[derive(Serialize)]
struct Summary {
    files: Vec<FileResult>,
}

struct UsxState {
    book_code: String,
    current_chapter: String,
    current_verse: String,
    current_plain: String,
    current_styled: String,
    current_footnotes: Vec<String>,
    current_crossrefs: Vec<String>,
    current_subtitle: String,
    rows: Vec<Row>,
}

fn main() {
    let args: Vec<String> = env::args().collect();
    let mut inputs: Vec<String> = Vec::new();
    let mut output_folder: Option<String> = None;
    let mut quiet = false;
    let mut json_out = false;

    let mut i = 1;
    while i < args.len() {
        match args[i].as_str() {
            "-h" | "--help" => {
                show_usage();
                return;
            }
            "-input" | "--input" => {
                i += 1;
                if i >= args.len() {
                    eprintln!("Missing value for -input");
                    std::process::exit(1);
                }
                inputs.push(args[i].clone());
            }
            "-output" | "--output" => {
                i += 1;
                if i >= args.len() {
                    eprintln!("Missing value for -output");
                    std::process::exit(1);
                }
                output_folder = Some(args[i].clone());
            }
            "-quiet" | "--quiet" => {
                quiet = true;
            }
            "-json" | "--json" => {
                json_out = true;
            }
            other => {
                inputs.push(other.to_string());
            }
        }
        i += 1;
    }

    if inputs.is_empty() {
        show_usage();
        return;
    }

    let input_items = match resolve_input_items(&inputs) {
        Ok(items) => items,
        Err(err) => {
            fail(&err, json_out);
        }
    };

    let files = match collect_files(&input_items) {
        Ok(items) => items,
        Err(err) => {
            fail(&err, json_out);
        }
    };

    if files.is_empty() {
        fail("No .usx, .usfm, or .sfm files found.", json_out);
    }

    if let Some(ref folder) = output_folder {
        if let Err(err) = fs::create_dir_all(folder) {
            fail(&err.to_string(), json_out);
        }
    }

    let mut summary = Summary { files: Vec::new() };

    for path in files {
        let ext = path.extension().and_then(|e| e.to_str()).unwrap_or("").to_lowercase();
        let csv_path = output_path(&path, output_folder.as_deref());

        let result = match ext.as_str() {
            "usx" => convert_usx_to_csv(&path, &csv_path, quiet).map(|rows| {
                summary.files.push(FileResult {
                    input: path.display().to_string(),
                    output: csv_path.display().to_string(),
                    format: "usx".to_string(),
                    rows,
                });
            }),
            "usfm" | "sfm" => convert_usfm_to_csv(&path, &csv_path, quiet).map(|rows| {
                summary.files.push(FileResult {
                    input: path.display().to_string(),
                    output: csv_path.display().to_string(),
                    format: ext.clone(),
                    rows,
                });
            }),
            _ => Ok(()),
        };

        if let Err(err) = result {
            fail(&err, json_out);
        }
    }

    if json_out {
        let json = serde_json::to_string_pretty(&summary).unwrap_or_else(|_| "{}".to_string());
        println!("{}", json);
        return;
    }

    println!("All conversions completed.");
}

fn show_usage() {
    println!("usxtocsv (Rust) - Convert USX/USFM/SFM to CSV");
    println!();
    println!("Usage:");
    println!("  usxtocsv -input <file|folder|wildcard> [-output <folder>]");
    println!("  usxtocsv -input <path1> -input <path2>");
    println!("  usxtocsv -quiet -json");
    println!("  usxtocsv --help");
}

fn resolve_input_items(inputs: &[String]) -> Result<Vec<PathBuf>, String> {
    let mut items: Vec<PathBuf> = Vec::new();

    for raw in inputs {
        for part in raw.split(',') {
            let p = part.trim();
            if p.is_empty() {
                continue;
            }

            if has_wildcard(p) {
                let mut matched = false;
                for entry in glob(p).map_err(|e| e.to_string())? {
                    let path = entry.map_err(|e| e.to_string())?;
                    items.push(path);
                    matched = true;
                }
                if !matched {
                    return Err(format!("Input path not found: {}", p));
                }
            } else {
                let path = PathBuf::from(p);
                if !path.exists() {
                    return Err(format!("Input path not found: {}", p));
                }
                items.push(path);
            }
        }
    }

    Ok(items)
}

fn collect_files(items: &[PathBuf]) -> Result<Vec<PathBuf>, String> {
    let mut files: Vec<PathBuf> = Vec::new();

    for item in items {
        let metadata = fs::metadata(item).map_err(|_| format!("Input path not found: {}", item.display()))?;

        if metadata.is_dir() {
            for entry in fs::read_dir(item).map_err(|e| e.to_string())? {
                let entry = entry.map_err(|e| e.to_string())?;
                let path = entry.path();
                if path.is_dir() {
                    continue;
                }
                if is_supported_ext(&path) {
                    files.push(path);
                }
            }
            continue;
        }

        if !is_supported_ext(item) {
            return Err("Input must be a .usx, .usfm, or .sfm file, or a folder containing them.".to_string());
        }
        files.push(item.clone());
    }

    Ok(files)
}

fn output_path(input_path: &Path, output_folder: Option<&str>) -> PathBuf {
    if let Some(folder) = output_folder {
        let base = input_path.file_stem().and_then(|s| s.to_str()).unwrap_or("output");
        return PathBuf::from(folder).join(format!("{}.csv", base));
    }

    let mut p = input_path.to_path_buf();
    p.set_extension("csv");
    p
}

fn is_supported_ext(path: &Path) -> bool {
    match path.extension().and_then(|e| e.to_str()).unwrap_or("").to_lowercase().as_str() {
        "usx" | "usfm" | "sfm" => true,
        _ => false,
    }
}

fn has_wildcard(text: &str) -> bool {
    text.contains('*') || text.contains('?') || text.contains('[')
}

fn convert_usx_to_csv(usx_path: &Path, csv_path: &Path, quiet: bool) -> Result<usize, String> {
    if !quiet {
        eprintln!("Processing (USX) {}", usx_path.display());
    }
    let root = parse_xml(usx_path).map_err(|e| e.to_string())?;
    if root.name != "usx" {
        return Err(format!("No <usx> root found in {}", usx_path.display()));
    }

    let book_node = find_first_child(&root, "book")
        .ok_or_else(|| format!("No <book> found in {}", usx_path.display()))?;

    let mut state = UsxState {
        book_code: get_attr_value(book_node, "code"),
        current_chapter: String::new(),
        current_verse: String::new(),
        current_plain: String::new(),
        current_styled: String::new(),
        current_footnotes: Vec::new(),
        current_crossrefs: Vec::new(),
        current_subtitle: String::new(),
        rows: Vec::new(),
    };

    for child in &root.children {
        process_usx_node(child, &mut state);
    }

    sort_rows(&mut state.rows);
    write_csv(csv_path, &state.rows).map_err(|e| e.to_string())?;
    if !quiet {
        eprintln!("Created CSV: {}", csv_path.display());
    }
    Ok(state.rows.len())
}

fn convert_usfm_to_csv(usfm_path: &Path, csv_path: &Path, quiet: bool) -> Result<usize, String> {
    if !quiet {
        eprintln!("Processing (USFM/SFM) {}", usfm_path.display());
    }
    let data = fs::read_to_string(usfm_path).map_err(|e| e.to_string())?;
    let normalized = data.replace("\r\n", "\n");
    let lines: Vec<&str> = normalized.split('\n').collect();

    let mut book_code = usfm_path
        .file_stem()
        .and_then(|s| s.to_str())
        .unwrap_or("UNKNOWN")
        .to_string();

    let re_id = Regex::new(r"(?i)^\\id\s+(\S+)").unwrap();
    for line in &lines {
        let l = line.trim();
        if l.is_empty() {
            continue;
        }
        if let Some(caps) = re_id.captures(l) {
            if let Some(val) = caps.get(1) {
                book_code = val.as_str().to_string();
                break;
            }
        }
    }

    let re_chapter = Regex::new(r"(?i)^\\c\s+(\d+)\b").unwrap();
    let re_heading = Regex::new(r"(?i)^\\(s[0-3]?|sp|ms|mr|mt[12]?)\s*(.*)$").unwrap();
    let re_verse = Regex::new(r"(?i)^\\v\s+(\d+)\s*(.*)$").unwrap();
    let re_para = Regex::new(r"(?i)^\\(m|p|pi|q[0-4]?|qt[0-4]?)\s*(.*)$").unwrap();

    let mut rows: Vec<Row> = Vec::new();
    let mut current_chapter = String::new();
    let mut current_verse = String::new();
    let mut current_plain = String::new();
    let mut current_styled = String::new();
    let mut current_footnotes: Vec<String> = Vec::new();
    let mut current_crossrefs: Vec<String> = Vec::new();
    let mut current_subtitle = String::new();

    for line in lines {
        let l = line.trim();
        if l.is_empty() {
            continue;
        }

        if let Some(caps) = re_chapter.captures(l) {
            if !current_verse.is_empty() {
                add_current_verse_usfm(
                    &mut rows,
                    &book_code,
                    &current_chapter,
                    &current_verse,
                    &current_plain,
                    &current_styled,
                    &current_footnotes,
                    &current_crossrefs,
                    &current_subtitle,
                );
            }
            current_verse.clear();
            current_plain.clear();
            current_styled.clear();
            current_footnotes.clear();
            current_crossrefs.clear();

            current_chapter = caps.get(1).unwrap().as_str().to_string();
            continue;
        }

        if let Some(caps) = re_heading.captures(l) {
            let mut head_text = caps.get(2).map(|m| m.as_str()).unwrap_or("").to_string();
            head_text = extract_notes_from_usfm_segment(&head_text, &mut current_footnotes, &mut current_crossrefs);
            let re_unknown = Regex::new(r"(?i)\\\+?[a-z0-9]+\*?").unwrap();
            head_text = re_unknown.replace_all(&head_text, " ").to_string();
            head_text = normalize_whitespace(&head_text);
            if !head_text.is_empty() {
                current_subtitle = head_text;
            }
            continue;
        }

        if let Some(caps) = re_verse.captures(l) {
            if !current_verse.is_empty() {
                add_current_verse_usfm(
                    &mut rows,
                    &book_code,
                    &current_chapter,
                    &current_verse,
                    &current_plain,
                    &current_styled,
                    &current_footnotes,
                    &current_crossrefs,
                    &current_subtitle,
                );
            }

            current_verse = caps.get(1).unwrap().as_str().to_string();
            current_plain.clear();
            current_styled.clear();
            current_footnotes.clear();
            current_crossrefs.clear();

            let rest = caps.get(2).map(|m| m.as_str()).unwrap_or("");
            if !rest.is_empty() {
                process_usfm_content_segment(
                    rest,
                    &mut current_plain,
                    &mut current_styled,
                    &mut current_footnotes,
                    &mut current_crossrefs,
                );
            }
            continue;
        }

        if let Some(caps) = re_para.captures(l) {
            let rest = caps.get(2).map(|m| m.as_str()).unwrap_or("");
            if !current_verse.is_empty() && !rest.is_empty() {
                process_usfm_content_segment(
                    rest,
                    &mut current_plain,
                    &mut current_styled,
                    &mut current_footnotes,
                    &mut current_crossrefs,
                );
            }
            continue;
        }

        if !current_verse.is_empty() {
            process_usfm_content_segment(
                l,
                &mut current_plain,
                &mut current_styled,
                &mut current_footnotes,
                &mut current_crossrefs,
            );
        }
    }

    if !current_verse.is_empty() {
        add_current_verse_usfm(
            &mut rows,
            &book_code,
            &current_chapter,
            &current_verse,
            &current_plain,
            &current_styled,
            &current_footnotes,
            &current_crossrefs,
            &current_subtitle,
        );
    }

    sort_rows(&mut rows);
    write_csv(csv_path, &rows).map_err(|e| e.to_string())?;
    if !quiet {
        eprintln!("Created CSV: {}", csv_path.display());
    }
    Ok(rows.len())
}

fn add_current_verse_usfm(
    rows: &mut Vec<Row>,
    book: &str,
    chapter: &str,
    verse: &str,
    plain: &str,
    styled: &str,
    footnotes: &[String],
    crossrefs: &[String],
    subtitle: &str,
) {
    let plain_trim = plain.trim();
    let styled_trim = styled.trim();
    let subtitle_trim = subtitle.trim();

    if !book.is_empty() && !chapter.is_empty() && !verse.is_empty() && !plain_trim.is_empty() {
        rows.push(Row {
            book: book.to_string(),
            chapter: chapter.to_string(),
            verse: verse.to_string(),
            text_plain: plain_trim.to_string(),
            text_styled: styled_trim.to_string(),
            footnotes: footnotes.join(" | "),
            crossrefs: crossrefs.join(" | "),
            subtitle: subtitle_trim.to_string(),
        });
    }
}

fn process_usfm_content_segment(
    segment: &str,
    current_plain: &mut String,
    current_styled: &mut String,
    footnotes: &mut Vec<String>,
    crossrefs: &mut Vec<String>,
) {
    if segment.trim().is_empty() {
        return;
    }

    let re_sup = Regex::new(r"(?is)\\\+?sup\b.*?\\\+?sup\*").unwrap();
    let mut seg = re_sup.replace_all(segment, " ").to_string();

    seg = extract_notes_from_usfm_segment(&seg, footnotes, crossrefs);
    if seg.trim().is_empty() {
        return;
    }

    let style_map = vec![
        ("wj", "wj"),
        ("add", "add"),
        ("nd", "nd"),
        ("it", "i"),
        ("bd", "b"),
        ("bdit", "bdit"),
    ];

    let mut styled = seg.clone();
    for (key, tag) in style_map {
        let re_open = Regex::new(&format!(r"(?i)\\\+?{}\b\s*", key)).unwrap();
        let re_close = Regex::new(&format!(r"(?i)\\\+?{}\*\s*", key)).unwrap();
        styled = re_open.replace_all(&styled, format!("<{}>", tag)).to_string();
        styled = re_close.replace_all(&styled, format!("</{}>", tag)).to_string();
    }

    let re_unknown = Regex::new(r"(?i)\\\+?[a-z0-9]+\*?").unwrap();
    styled = normalize_whitespace(&re_unknown.replace_all(&styled, " ").to_string());
    let plain = normalize_whitespace(&re_unknown.replace_all(&seg, " ").to_string());

    if !plain.is_empty() {
        if !current_plain.is_empty() {
            current_plain.push(' ');
            current_styled.push(' ');
        }
        current_plain.push_str(&plain);
        current_styled.push_str(&styled);
    }
}

fn extract_notes_from_usfm_segment(
    segment: &str,
    footnotes: &mut Vec<String>,
    crossrefs: &mut Vec<String>,
) -> String {
    if segment.trim().is_empty() {
        return segment.to_string();
    }

    let re_foot = Regex::new(r"(?is)\\f\b(.*?\\f\*)").unwrap();
    let text = re_foot
        .replace_all(segment, |caps: &Captures| {
            let full = format!("\\f{}", &caps[1]);
            let ft = extract_ft_from_usfm_note_text(&full);
            if !ft.is_empty() {
                footnotes.push(ft);
            }
            " "
        })
        .to_string();

    let re_cross = Regex::new(r"(?is)\\x\b(.*?\\x\*)").unwrap();
    re_cross
        .replace_all(&text, |caps: &Captures| {
            let full = format!("\\x{}", &caps[1]);
            let ft = extract_ft_from_usfm_note_text(&full);
            if !ft.is_empty() {
                crossrefs.push(ft);
            }
            " "
        })
        .to_string()
}

fn extract_ft_from_usfm_note_text(note_text: &str) -> String {
    if note_text.trim().is_empty() {
        return String::new();
    }

    let re = Regex::new(r"(?i)\\ft\b([^\\]*)").unwrap();
    if let Some(caps) = re.captures(note_text) {
        if let Some(m) = caps.get(1) {
            return normalize_whitespace(m.as_str());
        }
    }

    String::new()
}

fn parse_xml(path: &Path) -> Result<Node, io::Error> {
    let mut reader = Reader::from_file(path)
        .map_err(|e| io::Error::new(io::ErrorKind::Other, e))?;
    reader.trim_text(false);

    let mut buf = Vec::new();
    let mut stack: Vec<Node> = Vec::new();
    let mut root: Option<Node> = None;

    loop {
        match reader.read_event_into(&mut buf) {
            Ok(Event::Start(e)) => {
                let mut attrs = HashMap::new();
                for attr in e.attributes().flatten() {
                    let key = String::from_utf8_lossy(attr.key.as_ref()).to_string();
                    let value = attr.unescape_value().unwrap_or_default().to_string();
                    attrs.insert(key, value);
                }
                let node = Node {
                    node_type: NodeType::Element,
                    name: String::from_utf8_lossy(e.name().as_ref()).to_string(),
                    attrs,
                    children: Vec::new(),
                    text: String::new(),
                };
                stack.push(node);
            }
            Ok(Event::End(_e)) => {
                if let Some(node) = stack.pop() {
                    if let Some(parent) = stack.last_mut() {
                        parent.children.push(node);
                    } else {
                        root = Some(node);
                    }
                }
            }
            Ok(Event::Text(e)) => {
                if let Some(parent) = stack.last_mut() {
                    let text = e.unescape().unwrap_or_default().to_string();
                    if !text.is_empty() {
                        parent.children.push(Node {
                            node_type: NodeType::Text,
                            name: String::new(),
                            attrs: HashMap::new(),
                            children: Vec::new(),
                            text,
                        });
                    }
                }
            }
            Ok(Event::Eof) => break,
            Err(e) => return Err(io::Error::new(io::ErrorKind::Other, e)),
            _ => {}
        }
        buf.clear();
    }

    Ok(root.unwrap_or(Node {
        node_type: NodeType::Element,
        name: String::new(),
        attrs: HashMap::new(),
        children: Vec::new(),
        text: String::new(),
    }))
}

fn process_usx_node(node: &Node, state: &mut UsxState) {
    match node.node_type {
        NodeType::Element => match node.name.as_str() {
            "chapter" => {
                state.current_chapter = get_attr_value(node, "number");
            }
            "verse" => {
                let sid = get_attr_value(node, "sid");
                let eid = get_attr_value(node, "eid");
                if !sid.is_empty() {
                    state.current_verse = get_attr_value(node, "number");
                    state.current_plain.clear();
                    state.current_styled.clear();
                    state.current_footnotes.clear();
                    state.current_crossrefs.clear();
                } else if !eid.is_empty() {
                    add_current_verse_usx(state);
                    state.current_verse.clear();
                    state.current_plain.clear();
                    state.current_styled.clear();
                    state.current_footnotes.clear();
                    state.current_crossrefs.clear();
                }
            }
            "note" => {
                process_usx_note(node, state);
                return;
            }
            "para" => {
                let style = get_attr_value(node, "style");
                if is_subtitle_style(&style) {
                    let subtitle = normalize_whitespace(&inner_text(node));
                    if !subtitle.is_empty() {
                        state.current_subtitle = subtitle;
                    }
                }
            }
            "char" => {
                let style = get_attr_value(node, "style");
                if style == "sup" {
                    return;
                }
                let mut tag = String::new();
                if !style.is_empty() {
                    tag = get_styled_tag_name(&style);
                }
                if !state.current_verse.is_empty() && !tag.is_empty() {
                    state.current_styled.push_str(&format!("<{}>", tag));
                }
                for child in &node.children {
                    process_usx_node(child, state);
                }
                if !state.current_verse.is_empty() && !tag.is_empty() {
                    state.current_styled.push_str(&format!("</{}>", tag));
                }
                return;
            }
            _ => {}
        },
        NodeType::Text => {
            if state.current_verse.is_empty() {
                return;
            }
            let text = normalize_whitespace(&node.text);
            if text.is_empty() {
                return;
            }
            if !state.current_plain.is_empty() {
                state.current_plain.push(' ');
                state.current_styled.push(' ');
            }
            state.current_plain.push_str(&text);
            state.current_styled.push_str(&text);
        }
    }

    for child in &node.children {
        process_usx_node(child, state);
    }
}

fn add_current_verse_usx(state: &mut UsxState) {
    let plain = state.current_plain.trim().to_string();
    let styled = state.current_styled.trim().to_string();
    let subtitle = state.current_subtitle.trim().to_string();

    if !state.book_code.is_empty()
        && !state.current_chapter.is_empty()
        && !state.current_verse.is_empty()
        && !plain.is_empty()
    {
        state.rows.push(Row {
            book: state.book_code.clone(),
            chapter: state.current_chapter.clone(),
            verse: state.current_verse.clone(),
            text_plain: plain,
            text_styled: styled,
            footnotes: state.current_footnotes.join(" | "),
            crossrefs: state.current_crossrefs.join(" | "),
            subtitle,
        });
    }
}

fn process_usx_note(node: &Node, state: &mut UsxState) {
    let style = get_attr_value(node, "style");
    let ft = extract_ft_from_note(node);
    if ft.is_empty() {
        return;
    }
    if style.starts_with('x') {
        state.current_crossrefs.push(ft);
    } else {
        state.current_footnotes.push(ft);
    }
}

fn extract_ft_from_note(node: &Node) -> String {
    if let Some(ft_node) = find_ft_node(node) {
        let raw = inner_text(&ft_node);
        return normalize_whitespace(&raw);
    }
    String::new()
}

fn find_ft_node(node: &Node) -> Option<Node> {
    if node.node_type == NodeType::Element && node.name == "char" {
        if get_attr_value(node, "style") == "ft" {
            return Some(node.clone());
        }
    }
    for child in &node.children {
        if let Some(found) = find_ft_node(child) {
            return Some(found);
        }
    }
    None
}

fn inner_text(node: &Node) -> String {
    match node.node_type {
        NodeType::Text => node.text.clone(),
        NodeType::Element => {
            let mut out = String::new();
            for child in &node.children {
                out.push_str(&inner_text(child));
            }
            out
        }
    }
}

fn find_first_child<'a>(node: &'a Node, name: &str) -> Option<&'a Node> {
    node.children
        .iter()
        .find(|child| matches!(child.node_type, NodeType::Element) && child.name == name)
}

fn get_attr_value(node: &Node, name: &str) -> String {
    node.attrs.get(name).cloned().unwrap_or_default()
}

fn normalize_whitespace(text: &str) -> String {
    if text.trim().is_empty() {
        return String::new();
    }
    text.split_whitespace().collect::<Vec<_>>().join(" ")
}

fn get_styled_tag_name(style: &str) -> String {
    match style {
        "wj" => "wj",
        "add" => "add",
        "nd" => "nd",
        "bdit" => "bdit",
        "it" => "i",
        "bd" => "b",
        _ => "span",
    }
    .to_string()
}

fn is_subtitle_style(style: &str) -> bool {
    matches!(
        style,
        "s" | "s1" | "s2" | "s3" | "sp" | "ms" | "mr" | "mt" | "mt1" | "mt2"
    )
}

fn sort_rows(rows: &mut Vec<Row>) {
    rows.sort_by(|a, b| {
        let book_cmp = a.book.cmp(&b.book);
        if book_cmp != std::cmp::Ordering::Equal {
            return book_cmp;
        }
        let ca = a.chapter.parse::<i32>().unwrap_or(0);
        let cb = b.chapter.parse::<i32>().unwrap_or(0);
        if ca != cb {
            return ca.cmp(&cb);
        }
        a.verse.cmp(&b.verse)
    });
}

fn write_csv(path: &Path, rows: &[Row]) -> Result<(), io::Error> {
    let mut writer = Writer::from_path(path)?;
    writer.write_record([
        "Book",
        "Chapter",
        "Verse",
        "TextPlain",
        "TextStyled",
        "Footnotes",
        "Crossrefs",
        "Subtitle",
    ])?;
    for row in rows {
        writer.write_record([
            &row.book,
            &row.chapter,
            &row.verse,
            &row.text_plain,
            &row.text_styled,
            &row.footnotes,
            &row.crossrefs,
            &row.subtitle,
        ])?;
    }
    writer.flush()?;
    Ok(())
}

fn fail(message: &str, json_out: bool) -> ! {
    if json_out {
        let out = serde_json::json!({ "error": message });
        println!("{}", serde_json::to_string_pretty(&out).unwrap_or_else(|_| "{}".to_string()));
        std::process::exit(1);
    }
    eprintln!("{}", message);
    std::process::exit(1);
}
