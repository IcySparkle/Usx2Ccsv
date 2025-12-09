param(
    [Parameter(Mandatory = $true)]
    [string]$InputPath,

    [Parameter(Mandatory = $false)]
    [string]$OutputFolder
)

# ---- Validate Input Path ----
if (-not (Test-Path $InputPath)) {
    Write-Error "Input path not found: $InputPath"
    exit 1
}

# ---- Prepare Output Folder ----
if ($OutputFolder) {
    if (-not (Test-Path $OutputFolder)) {
        New-Item -ItemType Directory -Path $OutputFolder | Out-Null
        Write-Host "Created output folder: $OutputFolder"
    }
}

# ---- Collect USX Files ----
$usxFiles = @()

if ((Get-Item $InputPath).PSIsContainer) {
    # Folder → all *.usx
    $usxFiles = Get-ChildItem -Path $InputPath -Filter *.usx
}
else {
    # Single file
    if ($InputPath.ToLower().EndsWith(".usx")) {
        $usxFiles = ,(Get-Item $InputPath)
    }
    else {
        Write-Error "Input must be a .usx file or a folder containing .usx files."
        exit 1
    }
}

if ($usxFiles.Count -eq 0) {
    Write-Error "No .usx files found."
    exit 1
}

# ---- Helper: get attribute value safely ----
function Get-AttrValue {
    param(
        [Parameter(Mandatory = $true)] $Node,
        [Parameter(Mandatory = $true)] [string]$Name
    )
    if (-not $Node.Attributes) { return $null }
    $attr = $Node.Attributes[$Name]
    if ($attr) { return $attr.Value } else { return $null }
}

# ---- Map USX char styles to simple tags ----
function Get-StyledTagName {
    param(
        [string]$style
    )
    switch ($style) {
        'wj'   { 'wj' }   # Words of Jesus
        'add'  { 'add' }  # Additions
        'nd'   { 'nd' }   # Divine name
        'bdit' { 'bdit' } # Bold-italic
        'it'   { 'i' }    # Italic
        'bd'   { 'b' }    # Bold
        default { 'span' } # Fallback; change to $null if you prefer no tag
    }
}

# ---- Subtitle helper: extract plain inner text ----
function Get-PlainInnerText {
    param(
        [System.Xml.XmlNode]$node
    )
    if (-not $node) { return "" }

    $raw = $node.InnerText
    if ([string]::IsNullOrWhiteSpace($raw)) { return "" }

    $clean = ($raw -replace "\s+", " ").Trim()
    return $clean
}

# ---- Find first <char style="ft"> node under a note ----
function Find-FtNode {
    param(
        [System.Xml.XmlNode]$node
    )

    if (-not $node) { return $null }

    if ($node.LocalName -eq 'char') {
        $style = Get-AttrValue -Node $node -Name 'style'
        if ($style -eq 'ft') {
            return $node
        }
    }

    foreach ($child in $node.ChildNodes) {
        $result = Find-FtNode -node $child
        if ($result) { return $result }
    }

    return $null
}

# ---- Extract only the FT text from a note ----
function ExtractFTFromNote {
    param(
        [System.Xml.XmlNode]$noteNode
    )

    $ftNode = Find-FtNode -node $noteNode
    if (-not $ftNode) { return "" }

    $raw = $ftNode.InnerText
    if ([string]::IsNullOrWhiteSpace($raw)) { return "" }

    $clean = ($raw -replace "\s+", " ").Trim()
    return $clean
}

function Convert-UsxToCsv {
    param(
        [string]$UsxPath,
        [string]$CsvPath
    )

    Write-Host "Processing: $UsxPath"

    [xml]$doc = Get-Content -LiteralPath $UsxPath -Encoding UTF8
    
    $nsUri = $doc.DocumentElement.NamespaceURI
    $nsMgr = New-Object System.Xml.XmlNamespaceManager($doc.NameTable)
    $nsMgr.AddNamespace("u", $nsUri)

    $bookNode = $doc.SelectSingleNode("/u:usx/u:book", $nsMgr)
    if (-not $bookNode) {
        Write-Error "No <book> found in $UsxPath"
        return
    }
    $bookCode = Get-AttrValue -Node $bookNode -Name "code"

    # Per-verse state
    $script:currentChapter     = $null
    $script:currentVerse       = $null
    $script:currentPlainText   = ""
    $script:currentStyledText  = ""
    $script:currentFootnotes   = New-Object System.Collections.Generic.List[string]
    $script:currentCrossrefs   = New-Object System.Collections.Generic.List[string]
    $script:currentSubtitle    = ""   # last seen subtitle text (section heading)

    $rows = New-Object System.Collections.Generic.List[object]

    function Add-CurrentVerse {
        param($Book, $Chapter, $Verse)

        $plain   = $script:currentPlainText.Trim()
        $styled  = $script:currentStyledText.Trim()
        $subText = $script:currentSubtitle.Trim()

        if ($Book -and $Chapter -and $Verse -and $plain) {
            $rows.Add([pscustomobject]@{
                Book       = $Book
                Chapter    = $Chapter
                Verse      = $Verse
                TextPlain  = $plain
                TextStyled = $styled
                Footnotes  = ($script:currentFootnotes -join " | ")
                Crossrefs  = ($script:currentCrossrefs -join " | ")
                Subtitle   = $subText
            })
        }
    }

    function Process-NoteNode {
        param(
            [System.Xml.XmlNode]$noteNode
        )

        $style = Get-AttrValue -Node $noteNode -Name "style"
        $ft    = ExtractFTFromNote -noteNode $noteNode
        if (-not $ft) { return }

        # style starting with 'x' = cross-reference; else treat as footnote
        if ($style -and $style.StartsWith("x")) {
            $script:currentCrossrefs.Add($ft)
        }
        else {
            $script:currentFootnotes.Add($ft)
        }
    }

    # Helper to decide if a para style is a subtitle/heading
    function Is-SubtitleStyle {
        param([string]$style)

        if (-not $style) { return $false }

        # Basic heading styles
        $subtitleStyles = @(
            's','s1','s2','s3','sp',
            'ms','mr',
            'mt','mt1','mt2'
        )

        return $subtitleStyles -contains $style
    }

    function Process-Node {
        param(
            [System.Xml.XmlNode]$node,
            [System.Xml.XmlNamespaceManager]$nsMgr
        )

        switch ($node.NodeType) {
            'Element' {
                switch ($node.LocalName) {

                    'chapter' {
                        $script:currentChapter  = Get-AttrValue -Node $node -Name "number"
                        # If you want subtitles to reset at new chapter, uncomment:
                        # $script:currentSubtitle = ""
                    }

                    'verse' {
                        $sid = Get-AttrValue -Node $node -Name "sid"
                        $eid = Get-AttrValue -Node $node -Name "eid"

                        if ($sid) {
                            # Start of verse
                            $script:currentVerse       = Get-AttrValue -Node $node -Name "number"
                            $script:currentPlainText   = ""
                            $script:currentStyledText  = ""
                            $script:currentFootnotes.Clear()
                            $script:currentCrossrefs.Clear()
                        }
                        elseif ($eid) {
                            # End of verse
                            Add-CurrentVerse -Book $bookCode `
                                             -Chapter $script:currentChapter `
                                             -Verse $script:currentVerse

                            $script:currentVerse       = $null
                            $script:currentPlainText   = ""
                            $script:currentStyledText  = ""
                            $script:currentFootnotes.Clear()
                            $script:currentCrossrefs.Clear()
                        }
                    }

                    'note' {
                        Process-NoteNode -noteNode $node
                        return
                    }

                    'para' {
                        $style = Get-AttrValue -Node $node -Name "style"

                        # If this para is a heading/subtitle, capture it
                        if (Is-SubtitleStyle -style $style) {
                            $subtitleText = Get-PlainInnerText -node $node
                            if ($subtitleText) {
                                $script:currentSubtitle = $subtitleText
                            }
                        }

                        # Still process child nodes (verses etc.)
                        foreach ($child in $node.ChildNodes) {
                            Process-Node -node $child -nsMgr $nsMgr
                        }
                    }

                    'char' {
                        # Inline styled span
                        $style = Get-AttrValue -Node $node -Name "style"
                        $tag   = $null
                        if ($style) {
                            $tag = Get-StyledTagName -style $style
                        }

                        if ($script:currentVerse -and $tag) {
                            $script:currentStyledText += "<$tag>"
                        }

                        foreach ($child in $node.ChildNodes) {
                            Process-Node -node $child -nsMgr $nsMgr
                        }

                        if ($script:currentVerse -and $tag) {
                            $script:currentStyledText += "</$tag>"
                        }
                    }

                    default {
                        foreach ($child in $node.ChildNodes) {
                            Process-Node -node $child -nsMgr $nsMgr
                        }
                    }
                }
            }

            'Text' {
                if ($script:currentVerse) {
                    $t = $node.Value
                    if (-not [string]::IsNullOrWhiteSpace($t)) {
                        $t = $t.Trim()
                        if ($script:currentPlainText.Length -gt 0) {
                            $script:currentPlainText  += " "
                            $script:currentStyledText += " "
                        }
                        $script:currentPlainText  += $t
                        $script:currentStyledText += $t
                    }
                }
            }

            default {
                # Ignore comments, etc.
            }
        }
    }

    $root = $doc.SelectSingleNode("/u:usx", $nsMgr)
    foreach ($child in $root.ChildNodes) {
        Process-Node -node $child -nsMgr $nsMgr
    }

    # Sorting: chapters numeric, verses as string (works for a/b/c verses too)
    $rows |
        Sort-Object Book, {[int]$_.Chapter}, Verse |
        Export-Csv -Path $CsvPath -Encoding UTF8 -NoTypeInformation

    Write-Host "Created CSV: $CsvPath" -ForegroundColor Green
}

# ---- Run Conversion for All USX Files ----
foreach ($usx in $usxFiles) {

    if ($OutputFolder) {
        $csvPath = Join-Path $OutputFolder ($usx.BaseName + ".csv")
    } else {
        $csvPath = [System.IO.Path]::ChangeExtension($usx.FullName, ".csv")
    }

    Convert-UsxToCsv -UsxPath $usx.FullName -CsvPath $csvPath
}

Write-Host "All conversions completed."
