package parser

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
)

// XMLParser parses XML-formatted Liquibase changelogs.
type XMLParser struct{}

// xmlDatabaseChangeLog represents the root element of a Liquibase XML changelog.
type xmlDatabaseChangeLog struct {
	XMLName    xml.Name       `xml:"databaseChangeLog"`
	ChangeSets []xmlChangeSet `xml:"changeSet"`
	Includes   []xmlInclude   `xml:"include"`
	Properties []xmlProperty  `xml:"property"`
}

// xmlChangeSet represents a changeset element in XML.
type xmlChangeSet struct {
	Preconditions   *xmlPreconditions   `xml:"preConditions"`
	Rollback        *xmlRollback        `xml:"rollback"`
	ID              string              `xml:"id,attr"`
	Author          string              `xml:"author,attr"`
	Context         string              `xml:"context,attr"`
	Labels          string              `xml:"labels,attr"`
	DBMS            string              `xml:"dbms,attr"`
	RunAlways       string              `xml:"runAlways,attr"`
	RunOnChange     string              `xml:"runOnChange,attr"`
	FailOnError     string              `xml:"failOnError,attr"`
	LogicalFilePath string              `xml:"logicalFilePath,attr"`
	Comment         string              `xml:"comment"`
	DropTable       []xmlDropTable      `xml:"dropTable"`
	AddColumn       []xmlAddColumn      `xml:"addColumn"`
	DropColumn      []xmlDropColumn     `xml:"dropColumn"`
	CreateIndex     []xmlCreateIndex    `xml:"createIndex"`
	DropIndex       []xmlDropIndex      `xml:"dropIndex"`
	AddForeignKey   []xmlAddForeignKey  `xml:"addForeignKeyConstraint"`
	DropForeignKey  []xmlDropForeignKey `xml:"dropForeignKeyConstraint"`
	SQL             []xmlSQL            `xml:"sql"`
	Insert          []xmlInsert         `xml:"insert"`
	Update          []xmlUpdate         `xml:"update"`
	Delete          []xmlDelete         `xml:"delete"`
	CreateTable     []xmlCreateTable    `xml:"createTable"`
}

// xmlPreconditions represents preconditions element.
type xmlPreconditions struct {
	OnFail           string                `xml:"onFail,attr"`
	OnError          string                `xml:"onError,attr"`
	TableExists      []xmlTableExists      `xml:"tableExists"`
	ColumnExists     []xmlColumnExists     `xml:"columnExists"`
	IndexExists      []xmlIndexExists      `xml:"indexExists"`
	PrimaryKeyExists []xmlPrimaryKeyExists `xml:"primaryKeyExists"`
	ForeignKeyExists []xmlForeignKeyExists `xml:"foreignKeyConstraintExists"`
	SQLCheck         []xmlSQLCheck         `xml:"sqlCheck"`
}

// xmlTableExists represents a tableExists precondition.
type xmlTableExists struct {
	SchemaName string `xml:"schemaName,attr"`
	TableName  string `xml:"tableName,attr"`
}

// xmlColumnExists represents a columnExists precondition.
type xmlColumnExists struct {
	SchemaName string `xml:"schemaName,attr"`
	TableName  string `xml:"tableName,attr"`
	ColumnName string `xml:"columnName,attr"`
}

// xmlIndexExists represents an indexExists precondition.
type xmlIndexExists struct {
	SchemaName string `xml:"schemaName,attr"`
	TableName  string `xml:"tableName,attr"`
	IndexName  string `xml:"indexName,attr"`
}

// xmlPrimaryKeyExists represents a primaryKeyExists precondition.
type xmlPrimaryKeyExists struct {
	SchemaName     string `xml:"schemaName,attr"`
	TableName      string `xml:"tableName,attr"`
	PrimaryKeyName string `xml:"primaryKeyName,attr"`
}

// xmlForeignKeyExists represents a foreignKeyConstraintExists precondition.
type xmlForeignKeyExists struct {
	SchemaName     string `xml:"schemaName,attr"`
	ForeignKeyName string `xml:"foreignKeyName,attr"`
}

// xmlSQLCheck represents an sqlCheck precondition.
type xmlSQLCheck struct {
	ExpectedResult string `xml:"expectedResult,attr"`
	SQL            string `xml:",chardata"`
}

// xmlCreateTable represents a createTable change.
type xmlCreateTable struct {
	TableName  string      `xml:"tableName,attr"`
	SchemaName string      `xml:"schemaName,attr"`
	TableSpace string      `xml:"tablespace,attr"`
	Remarks    string      `xml:"remarks,attr"`
	Columns    []xmlColumn `xml:"column"`
}

// xmlColumn represents a column definition.
type xmlColumn struct {
	Constraints          *xmlConstraints `xml:"constraints"`
	Name                 string          `xml:"name,attr"`
	Type                 string          `xml:"type,attr"`
	Value                string          `xml:"value,attr"`
	DefaultValue         string          `xml:"defaultValue,attr"`
	DefaultValueComputed string          `xml:"defaultValueComputed,attr"`
	AutoIncrement        string          `xml:"autoIncrement,attr"`
	Remarks              string          `xml:"remarks,attr"`
}

// xmlConstraints represents column constraints.
type xmlConstraints struct {
	PrimaryKey     string `xml:"primaryKey,attr"`
	Nullable       string `xml:"nullable,attr"`
	Unique         string `xml:"unique,attr"`
	ForeignKeyName string `xml:"foreignKeyName,attr"`
	References     string `xml:"references,attr"`
	DeleteCascade  string `xml:"deleteCascade,attr"`
	PrimaryKeyName string `xml:"primaryKeyName,attr"`
}

// xmlDropTable represents a dropTable change.
type xmlDropTable struct {
	TableName          string `xml:"tableName,attr"`
	SchemaName         string `xml:"schemaName,attr"`
	CascadeConstraints string `xml:"cascadeConstraints,attr"`
}

// xmlAddColumn represents an addColumn change.
type xmlAddColumn struct {
	TableName  string      `xml:"tableName,attr"`
	SchemaName string      `xml:"schemaName,attr"`
	Columns    []xmlColumn `xml:"column"`
}

// xmlDropColumn represents a dropColumn change.
type xmlDropColumn struct {
	TableName  string `xml:"tableName,attr"`
	SchemaName string `xml:"schemaName,attr"`
	ColumnName string `xml:"columnName,attr"`
}

// xmlCreateIndex represents a createIndex change.
type xmlCreateIndex struct {
	IndexName  string      `xml:"indexName,attr"`
	TableName  string      `xml:"tableName,attr"`
	SchemaName string      `xml:"schemaName,attr"`
	Unique     string      `xml:"unique,attr"`
	TableSpace string      `xml:"tablespace,attr"`
	Columns    []xmlColumn `xml:"column"`
}

// xmlDropIndex represents a dropIndex change.
type xmlDropIndex struct {
	IndexName  string `xml:"indexName,attr"`
	TableName  string `xml:"tableName,attr"`
	SchemaName string `xml:"schemaName,attr"`
}

// xmlAddForeignKey represents an addForeignKeyConstraint change.
type xmlAddForeignKey struct {
	ConstraintName        string `xml:"constraintName,attr"`
	BaseTableName         string `xml:"baseTableName,attr"`
	BaseColumnNames       string `xml:"baseColumnNames,attr"`
	ReferencedTableName   string `xml:"referencedTableName,attr"`
	ReferencedColumnNames string `xml:"referencedColumnNames,attr"`
	OnDelete              string `xml:"onDelete,attr"`
	OnUpdate              string `xml:"onUpdate,attr"`
}

// xmlDropForeignKey represents a dropForeignKeyConstraint change.
type xmlDropForeignKey struct {
	ConstraintName string `xml:"constraintName,attr"`
	BaseTableName  string `xml:"baseTableName,attr"`
}

// xmlSQL represents a sql change.
type xmlSQL struct {
	DBMS            string `xml:"dbms,attr"`
	EndDelimiter    string `xml:"endDelimiter,attr"`
	SplitStatements string `xml:"splitStatements,attr"`
	StripComments   string `xml:"stripComments,attr"`
	SQL             string `xml:",chardata"`
}

// xmlInsert represents an insert change.
type xmlInsert struct {
	TableName  string      `xml:"tableName,attr"`
	SchemaName string      `xml:"schemaName,attr"`
	Columns    []xmlColumn `xml:"column"`
}

// xmlUpdate represents an update change.
type xmlUpdate struct {
	TableName  string      `xml:"tableName,attr"`
	SchemaName string      `xml:"schemaName,attr"`
	Where      string      `xml:"where"`
	Columns    []xmlColumn `xml:"column"`
}

// xmlDelete represents a delete change.
type xmlDelete struct {
	TableName  string `xml:"tableName,attr"`
	SchemaName string `xml:"schemaName,attr"`
	Where      string `xml:"where"`
}

// xmlRollback represents rollback instructions.
type xmlRollback struct {
	SQL         []xmlSQL         `xml:"sql"`
	DropTable   []xmlDropTable   `xml:"dropTable"`
	CreateTable []xmlCreateTable `xml:"createTable"`
	DropIndex   []xmlDropIndex   `xml:"dropIndex"`
	CreateIndex []xmlCreateIndex `xml:"createIndex"`
	DropColumn  []xmlDropColumn  `xml:"dropColumn"`
	AddColumn   []xmlAddColumn   `xml:"addColumn"`
}

// xmlInclude represents an include element.
type xmlInclude struct {
	File string `xml:"file,attr"`
}

// xmlProperty represents a property element.
type xmlProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

// Parse parses an XML changelog file.
func (p *XMLParser) Parse(filePath string) (*Changelog, error) {
	// Read file
	//nolint:gosec // G304: File path is provided by user for parsing
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse XML
	var xmlDoc xmlDatabaseChangeLog
	if err := xml.Unmarshal(data, &xmlDoc); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	// Convert to internal representation
	changelog := &Changelog{
		FilePath:      filePath,
		Format:        FormatXML,
		ChangeSets:    make([]ChangeSet, 0, len(xmlDoc.ChangeSets)),
		IncludedFiles: []string{filePath}, // XML parser doesn't support includes recursively yet
	}

	for i := range xmlDoc.ChangeSets {
		cs := p.convertChangeSet(&xmlDoc.ChangeSets[i], filePath)
		changelog.ChangeSets = append(changelog.ChangeSets, cs)
	}

	return changelog, nil
}

// CanParse checks if this parser can handle the given file.
func (p *XMLParser) CanParse(filePath string) bool {
	return DetectFormat(filePath) == FormatXML
}

// convertChangeSet converts an XML changeset to internal representation.
func (p *XMLParser) convertChangeSet(xmlCS *xmlChangeSet, filePath string) ChangeSet {
	cs := ChangeSet{
		ID:              xmlCS.ID,
		Author:          xmlCS.Author,
		FilePath:        filePath,
		Context:         xmlCS.Context,
		Comment:         xmlCS.Comment,
		LogicalFilePath: xmlCS.LogicalFilePath,
		Changes:         make([]Change, 0),
		Preconditions:   nil,
	}

	// Parse labels
	if xmlCS.Labels != "" {
		cs.Labels = strings.Split(xmlCS.Labels, ",")
		for i := range cs.Labels {
			cs.Labels[i] = strings.TrimSpace(cs.Labels[i])
		}
	}

	// Parse DBMS list
	if xmlCS.DBMS != "" {
		cs.DBMSList = strings.Split(xmlCS.DBMS, ",")
		for i := range cs.DBMSList {
			cs.DBMSList[i] = strings.TrimSpace(cs.DBMSList[i])
		}
	}

	// Parse boolean attributes
	cs.RunAlways = strings.EqualFold(xmlCS.RunAlways, "true")
	cs.RunOnChange = strings.EqualFold(xmlCS.RunOnChange, "true")
	cs.FailOnError = !strings.EqualFold(xmlCS.FailOnError, "false") // Default is true

	// Convert preconditions
	if xmlCS.Preconditions != nil {
		precond := p.convertPreconditions(xmlCS.Preconditions)
		if len(precond) > 0 {
			cs.Preconditions = &precond[0]
		}
	}

	// Convert changes
	cs.Changes = append(cs.Changes, p.convertCreateTables(xmlCS.CreateTable)...)
	cs.Changes = append(cs.Changes, p.convertDropTables(xmlCS.DropTable)...)
	cs.Changes = append(cs.Changes, p.convertAddColumns(xmlCS.AddColumn)...)
	cs.Changes = append(cs.Changes, p.convertDropColumns(xmlCS.DropColumn)...)
	cs.Changes = append(cs.Changes, p.convertCreateIndexes(xmlCS.CreateIndex)...)
	cs.Changes = append(cs.Changes, p.convertDropIndexes(xmlCS.DropIndex)...)
	cs.Changes = append(cs.Changes, p.convertAddForeignKeys(xmlCS.AddForeignKey)...)
	cs.Changes = append(cs.Changes, p.convertDropForeignKeys(xmlCS.DropForeignKey)...)
	cs.Changes = append(cs.Changes, p.convertSQLChanges(xmlCS.SQL)...)
	cs.Changes = append(cs.Changes, p.convertInserts(xmlCS.Insert)...)
	cs.Changes = append(cs.Changes, p.convertUpdates(xmlCS.Update)...)
	cs.Changes = append(cs.Changes, p.convertDeletes(xmlCS.Delete)...)

	// Convert rollback
	if xmlCS.Rollback != nil {
		cs.Rollback = p.convertRollback(xmlCS.Rollback)
	}

	return cs
}

// convertPreconditions converts XML preconditions to internal representation.
func (p *XMLParser) convertPreconditions(xmlPrec *xmlPreconditions) []Precondition {
	var preconditions []Precondition

	for _, te := range xmlPrec.TableExists {
		preconditions = append(preconditions, Precondition{
			Type:    "tableExists",
			OnFail:  xmlPrec.OnFail,
			OnError: xmlPrec.OnError,
			Attributes: map[string]string{
				"tableName":  te.TableName,
				"schemaName": te.SchemaName,
			},
		})
	}

	for _, ce := range xmlPrec.ColumnExists {
		preconditions = append(preconditions, Precondition{
			Type:    "columnExists",
			OnFail:  xmlPrec.OnFail,
			OnError: xmlPrec.OnError,
			Attributes: map[string]string{
				"tableName":  ce.TableName,
				"columnName": ce.ColumnName,
				"schemaName": ce.SchemaName,
			},
		})
	}

	for _, ie := range xmlPrec.IndexExists {
		preconditions = append(preconditions, Precondition{
			Type:    "indexExists",
			OnFail:  xmlPrec.OnFail,
			OnError: xmlPrec.OnError,
			Attributes: map[string]string{
				"tableName":  ie.TableName,
				"indexName":  ie.IndexName,
				"schemaName": ie.SchemaName,
			},
		})
	}

	return preconditions
}

// convertCreateTables converts createTable changes.
func (p *XMLParser) convertCreateTables(tables []xmlCreateTable) []Change {
	var changes []Change
	for _, table := range tables {
		changes = append(changes, Change{
			Type:      "createTable",
			TableName: table.TableName,
			Attributes: map[string]string{
				"tableName":  table.TableName,
				"schemaName": table.SchemaName,
				"remarks":    table.Remarks,
			},
		})
	}
	return changes
}

// convertDropTables converts dropTable changes.
func (p *XMLParser) convertDropTables(tables []xmlDropTable) []Change {
	var changes []Change
	for _, table := range tables {
		changes = append(changes, Change{
			Type:      "dropTable",
			TableName: table.TableName,
			Attributes: map[string]string{
				"tableName":          table.TableName,
				"schemaName":         table.SchemaName,
				"cascadeConstraints": table.CascadeConstraints,
			},
		})
	}
	return changes
}

// convertAddColumns converts addColumn changes.
func (p *XMLParser) convertAddColumns(columns []xmlAddColumn) []Change {
	var changes []Change
	for _, col := range columns {
		for _, c := range col.Columns {
			changes = append(changes, Change{
				Type:       "addColumn",
				TableName:  col.TableName,
				ColumnName: c.Name,
				Attributes: map[string]string{
					"tableName":  col.TableName,
					"columnName": c.Name,
					"type":       c.Type,
					"schemaName": col.SchemaName,
				},
			})
		}
	}
	return changes
}

// convertDropColumns converts dropColumn changes.
func (p *XMLParser) convertDropColumns(columns []xmlDropColumn) []Change {
	var changes []Change
	for _, col := range columns {
		changes = append(changes, Change{
			Type:       "dropColumn",
			TableName:  col.TableName,
			ColumnName: col.ColumnName,
			Attributes: map[string]string{
				"tableName":  col.TableName,
				"columnName": col.ColumnName,
				"schemaName": col.SchemaName,
			},
		})
	}
	return changes
}

// convertCreateIndexes converts createIndex changes.
func (p *XMLParser) convertCreateIndexes(indexes []xmlCreateIndex) []Change {
	var changes []Change
	for _, idx := range indexes {
		changes = append(changes, Change{
			Type:      "createIndex",
			TableName: idx.TableName,
			IndexName: idx.IndexName,
			Attributes: map[string]string{
				"indexName":  idx.IndexName,
				"tableName":  idx.TableName,
				"schemaName": idx.SchemaName,
				"unique":     idx.Unique,
			},
		})
	}
	return changes
}

// convertDropIndexes converts dropIndex changes.
func (p *XMLParser) convertDropIndexes(indexes []xmlDropIndex) []Change {
	var changes []Change
	for _, idx := range indexes {
		changes = append(changes, Change{
			Type:      "dropIndex",
			TableName: idx.TableName,
			IndexName: idx.IndexName,
			Attributes: map[string]string{
				"indexName":  idx.IndexName,
				"tableName":  idx.TableName,
				"schemaName": idx.SchemaName,
			},
		})
	}
	return changes
}

// convertAddForeignKeys converts addForeignKeyConstraint changes.
func (p *XMLParser) convertAddForeignKeys(fks []xmlAddForeignKey) []Change {
	var changes []Change
	for _, fk := range fks {
		changes = append(changes, Change{
			Type:      "addForeignKeyConstraint",
			TableName: fk.BaseTableName,
			Attributes: map[string]string{
				"constraintName":        fk.ConstraintName,
				"baseTableName":         fk.BaseTableName,
				"baseColumnNames":       fk.BaseColumnNames,
				"referencedTableName":   fk.ReferencedTableName,
				"referencedColumnNames": fk.ReferencedColumnNames,
				"onDelete":              fk.OnDelete,
				"onUpdate":              fk.OnUpdate,
			},
		})
	}
	return changes
}

// convertDropForeignKeys converts dropForeignKeyConstraint changes.
func (p *XMLParser) convertDropForeignKeys(fks []xmlDropForeignKey) []Change {
	var changes []Change
	for _, fk := range fks {
		changes = append(changes, Change{
			Type:      "dropForeignKeyConstraint",
			TableName: fk.BaseTableName,
			Attributes: map[string]string{
				"constraintName": fk.ConstraintName,
				"baseTableName":  fk.BaseTableName,
			},
		})
	}
	return changes
}

// convertSQLChanges converts sql changes.
func (p *XMLParser) convertSQLChanges(sqls []xmlSQL) []Change {
	var changes []Change
	for _, s := range sqls {
		changes = append(changes, Change{
			Type: "sql",
			SQL:  strings.TrimSpace(s.SQL),
			Attributes: map[string]string{
				"dbms":            s.DBMS,
				"endDelimiter":    s.EndDelimiter,
				"splitStatements": s.SplitStatements,
				"stripComments":   s.StripComments,
			},
		})
	}
	return changes
}

// convertInserts converts insert changes.
func (p *XMLParser) convertInserts(inserts []xmlInsert) []Change {
	var changes []Change
	for _, ins := range inserts {
		changes = append(changes, Change{
			Type:      "insert",
			TableName: ins.TableName,
			Attributes: map[string]string{
				"tableName":  ins.TableName,
				"schemaName": ins.SchemaName,
			},
		})
	}
	return changes
}

// convertUpdates converts update changes.
func (p *XMLParser) convertUpdates(updates []xmlUpdate) []Change {
	var changes []Change
	for _, upd := range updates {
		changes = append(changes, Change{
			Type:      "update",
			TableName: upd.TableName,
			Attributes: map[string]string{
				"tableName":  upd.TableName,
				"schemaName": upd.SchemaName,
				"where":      upd.Where,
			},
		})
	}
	return changes
}

// convertDeletes converts delete changes.
func (p *XMLParser) convertDeletes(deletes []xmlDelete) []Change {
	var changes []Change
	for _, del := range deletes {
		changes = append(changes, Change{
			Type:      "delete",
			TableName: del.TableName,
			Attributes: map[string]string{
				"tableName":  del.TableName,
				"schemaName": del.SchemaName,
				"where":      del.Where,
			},
		})
	}
	return changes
}

// convertRollback converts rollback instructions.
func (p *XMLParser) convertRollback(xmlRB *xmlRollback) *Rollback {
	rb := &Rollback{
		Changes: make([]Change, 0),
	}

	// Convert rollback changes
	rb.Changes = append(rb.Changes, p.convertCreateTables(xmlRB.CreateTable)...)
	rb.Changes = append(rb.Changes, p.convertDropTables(xmlRB.DropTable)...)
	rb.Changes = append(rb.Changes, p.convertCreateIndexes(xmlRB.CreateIndex)...)
	rb.Changes = append(rb.Changes, p.convertDropIndexes(xmlRB.DropIndex)...)
	rb.Changes = append(rb.Changes, p.convertAddColumns(xmlRB.AddColumn)...)
	rb.Changes = append(rb.Changes, p.convertDropColumns(xmlRB.DropColumn)...)

	// Convert rollback SQL
	if len(xmlRB.SQL) > 0 {
		var sqlParts []string
		for _, s := range xmlRB.SQL {
			sqlParts = append(sqlParts, strings.TrimSpace(s.SQL))
		}
		rb.SQL = strings.Join(sqlParts, "\n")
	}

	return rb
}
