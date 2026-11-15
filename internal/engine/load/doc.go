// Package load turns parsed data into the node tree a render draws from, and finds what each
// node reads. Its passes run in this order, the first two per file and the rest over one whole
// load, as engine's pipeline calls them:
//
//  1. Compile builds each node with its syntax: Compile, CompileStruct and CompileCategory in
//     node.go, a table in table.go, an option's datatype in datatype.go. format.go proves each
//     token's grammar and field paths as it goes, with path.go walking a path.
//  2. Bind gives each name its scope: bindNames in name.go, called by the compile.
//  3. Link ties a table to its parent table: LinkTable in table.go.
//  4. Resolve ties each reference and name to what it names, then compiles each format into ops:
//     CompileTemplates in resolve.go, whose doc lists its steps. It reads reference.go for a
//     reference, name.go for a name, read.go for an op's reads, and where.go, alphabetdraw.go,
//     adddays.go, reword.go and pathvalues.go for what one kind of token draws.
//  5. Settle fixes each record's columns: SettleRecords in record.go, with columnread.go finding
//     the column a template reads.
//
// Compile builds the nodes, and Link sets a table's siblings on the table. Bind, Resolve and
// Settle write what they find into the load's tree.LoadFinds, and read another pass's finds
// through it. LoadFinds.Finish then writes those into the nodes, and internal/engine/fence proves
// the finished tree. graph.go walks a node tree for any pass.
package load
