// Package parser provides functionality for parsing Liquibase changelog files.
package parser

import (
	"fmt"
)

// ConvertChange converts a change from a map[string]any structure to a Change object.
// This is used by YAML and JSON parsers which unmarshal to dynamic structures.
// Returns an error with full context (file path, changeset ID, change type, field) on validation failures.
func ConvertChange(changeType string, data any, filePath, changesetID string) (Change, error) {
	// Handle nil data
	if data == nil {
		return Change{}, fmt.Errorf("%s:changeSet[%s]:%s: change data is nil", filePath, changesetID, changeType)
	}

	// Convert data to map
	dataMap, ok := data.(map[string]any)
	if !ok {
		return Change{}, fmt.Errorf("%s:changeSet[%s]:%s: expected object, got %T", filePath, changesetID, changeType, data)
	}

	// Dispatch to type-specific converter
	switch changeType {
	case "createTable":
		return convertCreateTable(dataMap, filePath, changesetID)
	case "dropTable":
		return convertDropTable(dataMap, filePath, changesetID)
	case "addColumn":
		return convertAddColumn(dataMap, filePath, changesetID)
	case "dropColumn":
		return convertDropColumn(dataMap, filePath, changesetID)
	case "modifyDataType", "modifyColumn":
		return convertModifyColumn(dataMap, filePath, changesetID)
	case "renameColumn":
		return convertRenameColumn(dataMap, filePath, changesetID)
	case "renameTable":
		return convertRenameTable(dataMap, filePath, changesetID)
	case "addPrimaryKey":
		return convertAddPrimaryKey(dataMap, filePath, changesetID)
	case "dropPrimaryKey":
		return convertDropPrimaryKey(dataMap, filePath, changesetID)
	case "addForeignKeyConstraint":
		return convertAddForeignKey(dataMap, filePath, changesetID)
	case "dropForeignKeyConstraint":
		return convertDropForeignKey(dataMap, filePath, changesetID)
	case "addUniqueConstraint":
		return convertAddUniqueConstraint(dataMap, filePath, changesetID)
	case "dropUniqueConstraint":
		return convertDropUniqueConstraint(dataMap, filePath, changesetID)
	case "addNotNullConstraint":
		return convertAddNotNullConstraint(dataMap, filePath, changesetID)
	case "dropNotNullConstraint":
		return convertDropNotNullConstraint(dataMap, filePath, changesetID)
	case "addDefaultValue":
		return convertAddDefaultValue(dataMap, filePath, changesetID)
	case "dropDefaultValue":
		return convertDropDefaultValue(dataMap, filePath, changesetID)
	case "createIndex":
		return convertCreateIndex(dataMap, filePath, changesetID)
	case "dropIndex":
		return convertDropIndex(dataMap, filePath, changesetID)
	case "insert":
		return convertInsert(dataMap, filePath, changesetID)
	case "update":
		return convertUpdate(dataMap, filePath, changesetID)
	case "delete":
		return convertDelete(dataMap, filePath, changesetID)
	case "sql", "sqlFile":
		return convertSQL(dataMap, filePath, changesetID, changeType)
	default:
		// Unknown change type - create generic change
		return Change{
			Type:       changeType,
			Attributes: flattenMap(dataMap),
		}, nil
	}
}

// convertCreateTable converts a createTable change
func convertCreateTable(data map[string]any, filePath, changesetID string) (Change, error) {
	tableName, err := getRequiredString(data, "tableName", filePath, changesetID, "createTable")
	if err != nil {
		return Change{}, err
	}

	change := Change{
		Type:       "createTable",
		TableName:  tableName,
		Attributes: flattenMap(data),
	}

	// Extract columns if present (remove from attributes)
	if _, ok := data["columns"]; ok {
		delete(change.Attributes, "columns")
	}

	return change, nil
}

// convertDropTable converts a dropTable change
func convertDropTable(data map[string]any, filePath, changesetID string) (Change, error) {
	tableName, err := getRequiredString(data, "tableName", filePath, changesetID, "dropTable")
	if err != nil {
		return Change{}, err
	}

	return Change{
		Type:       "dropTable",
		TableName:  tableName,
		Attributes: flattenMap(data),
	}, nil
}

// convertAddColumn converts an addColumn change
func convertAddColumn(data map[string]any, filePath, changesetID string) (Change, error) {
	tableName, err := getRequiredString(data, "tableName", filePath, changesetID, "addColumn")
	if err != nil {
		return Change{}, err
	}

	change := Change{
		Type:       "addColumn",
		TableName:  tableName,
		Attributes: flattenMap(data),
	}

	// Extract column name if present in columns array
	if cols, ok := data["columns"]; ok {
		if colsArray, ok := cols.([]any); ok && len(colsArray) > 0 {
			if colMap, ok := colsArray[0].(map[string]any); ok {
				if colData, ok := colMap["column"].(map[string]any); ok {
					if name, ok := colData["name"].(string); ok {
						change.ColumnName = name
					}
				}
			}
		}
		delete(change.Attributes, "columns")
	}

	return change, nil
}

// convertDropColumn converts a dropColumn change
func convertDropColumn(data map[string]any, filePath, changesetID string) (Change, error) {
	tableName, err := getRequiredString(data, "tableName", filePath, changesetID, "dropColumn")
	if err != nil {
		return Change{}, err
	}

	columnName, err := getRequiredString(data, "columnName", filePath, changesetID, "dropColumn")
	if err != nil {
		return Change{}, err
	}

	return Change{
		Type:       "dropColumn",
		TableName:  tableName,
		ColumnName: columnName,
		Attributes: flattenMap(data),
	}, nil
}

// convertModifyColumn converts a modifyDataType or modifyColumn change
func convertModifyColumn(data map[string]any, filePath, changesetID string) (Change, error) {
	tableName, err := getRequiredString(data, "tableName", filePath, changesetID, "modifyColumn")
	if err != nil {
		return Change{}, err
	}

	columnName, err := getRequiredString(data, "columnName", filePath, changesetID, "modifyColumn")
	if err != nil {
		return Change{}, err
	}

	return Change{
		Type:       "modifyColumn",
		TableName:  tableName,
		ColumnName: columnName,
		Attributes: flattenMap(data),
	}, nil
}

// convertRenameColumn converts a renameColumn change
func convertRenameColumn(data map[string]any, filePath, changesetID string) (Change, error) {
	tableName, err := getRequiredString(data, "tableName", filePath, changesetID, "renameColumn")
	if err != nil {
		return Change{}, err
	}

	oldColumnName, err := getRequiredString(data, "oldColumnName", filePath, changesetID, "renameColumn")
	if err != nil {
		return Change{}, err
	}

	newColumnName, err := getRequiredString(data, "newColumnName", filePath, changesetID, "renameColumn")
	if err != nil {
		return Change{}, err
	}

	attrs := flattenMap(data)
	attrs["oldColumnName"] = oldColumnName
	attrs["newColumnName"] = newColumnName

	return Change{
		Type:       "renameColumn",
		TableName:  tableName,
		ColumnName: oldColumnName,
		Attributes: attrs,
	}, nil
}

// convertRenameTable converts a renameTable change
func convertRenameTable(data map[string]any, filePath, changesetID string) (Change, error) {
	oldTableName, err := getRequiredString(data, "oldTableName", filePath, changesetID, "renameTable")
	if err != nil {
		return Change{}, err
	}

	newTableName, err := getRequiredString(data, "newTableName", filePath, changesetID, "renameTable")
	if err != nil {
		return Change{}, err
	}

	attrs := flattenMap(data)
	attrs["newTableName"] = newTableName

	return Change{
		Type:       "renameTable",
		TableName:  oldTableName,
		Attributes: attrs,
	}, nil
}

// convertAddPrimaryKey converts an addPrimaryKey change
func convertAddPrimaryKey(data map[string]any, filePath, changesetID string) (Change, error) {
	tableName, err := getRequiredString(data, "tableName", filePath, changesetID, "addPrimaryKey")
	if err != nil {
		return Change{}, err
	}

	return Change{
		Type:       "addPrimaryKey",
		TableName:  tableName,
		Attributes: flattenMap(data),
	}, nil
}

// convertDropPrimaryKey converts a dropPrimaryKey change
func convertDropPrimaryKey(data map[string]any, filePath, changesetID string) (Change, error) {
	tableName, err := getRequiredString(data, "tableName", filePath, changesetID, "dropPrimaryKey")
	if err != nil {
		return Change{}, err
	}

	return Change{
		Type:       "dropPrimaryKey",
		TableName:  tableName,
		Attributes: flattenMap(data),
	}, nil
}

// convertAddForeignKey converts an addForeignKeyConstraint change
func convertAddForeignKey(data map[string]any, filePath, changesetID string) (Change, error) {
	baseTableName, err := getRequiredString(data, "baseTableName", filePath, changesetID, "addForeignKeyConstraint")
	if err != nil {
		return Change{}, err
	}

	return Change{
		Type:       "addForeignKey",
		TableName:  baseTableName,
		Attributes: flattenMap(data),
	}, nil
}

// convertDropForeignKey converts a dropForeignKeyConstraint change
func convertDropForeignKey(data map[string]any, filePath, changesetID string) (Change, error) {
	baseTableName, err := getRequiredString(data, "baseTableName", filePath, changesetID, "dropForeignKeyConstraint")
	if err != nil {
		return Change{}, err
	}

	return Change{
		Type:       "dropForeignKey",
		TableName:  baseTableName,
		Attributes: flattenMap(data),
	}, nil
}

// convertAddUniqueConstraint converts an addUniqueConstraint change
func convertAddUniqueConstraint(data map[string]any, filePath, changesetID string) (Change, error) {
	tableName, err := getRequiredString(data, "tableName", filePath, changesetID, "addUniqueConstraint")
	if err != nil {
		return Change{}, err
	}

	return Change{
		Type:       "addUniqueConstraint",
		TableName:  tableName,
		Attributes: flattenMap(data),
	}, nil
}

// convertDropUniqueConstraint converts a dropUniqueConstraint change
func convertDropUniqueConstraint(data map[string]any, filePath, changesetID string) (Change, error) {
	tableName, err := getRequiredString(data, "tableName", filePath, changesetID, "dropUniqueConstraint")
	if err != nil {
		return Change{}, err
	}

	return Change{
		Type:       "dropUniqueConstraint",
		TableName:  tableName,
		Attributes: flattenMap(data),
	}, nil
}

// convertAddNotNullConstraint converts an addNotNullConstraint change
func convertAddNotNullConstraint(data map[string]any, filePath, changesetID string) (Change, error) {
	tableName, err := getRequiredString(data, "tableName", filePath, changesetID, "addNotNullConstraint")
	if err != nil {
		return Change{}, err
	}

	columnName, err := getRequiredString(data, "columnName", filePath, changesetID, "addNotNullConstraint")
	if err != nil {
		return Change{}, err
	}

	return Change{
		Type:       "addNotNullConstraint",
		TableName:  tableName,
		ColumnName: columnName,
		Attributes: flattenMap(data),
	}, nil
}

// convertDropNotNullConstraint converts a dropNotNullConstraint change
func convertDropNotNullConstraint(data map[string]any, filePath, changesetID string) (Change, error) {
	tableName, err := getRequiredString(data, "tableName", filePath, changesetID, "dropNotNullConstraint")
	if err != nil {
		return Change{}, err
	}

	columnName, err := getRequiredString(data, "columnName", filePath, changesetID, "dropNotNullConstraint")
	if err != nil {
		return Change{}, err
	}

	return Change{
		Type:       "dropNotNullConstraint",
		TableName:  tableName,
		ColumnName: columnName,
		Attributes: flattenMap(data),
	}, nil
}

// convertAddDefaultValue converts an addDefaultValue change
func convertAddDefaultValue(data map[string]any, filePath, changesetID string) (Change, error) {
	tableName, err := getRequiredString(data, "tableName", filePath, changesetID, "addDefaultValue")
	if err != nil {
		return Change{}, err
	}

	columnName, err := getRequiredString(data, "columnName", filePath, changesetID, "addDefaultValue")
	if err != nil {
		return Change{}, err
	}

	return Change{
		Type:       "addDefaultValue",
		TableName:  tableName,
		ColumnName: columnName,
		Attributes: flattenMap(data),
	}, nil
}

// convertDropDefaultValue converts a dropDefaultValue change
func convertDropDefaultValue(data map[string]any, filePath, changesetID string) (Change, error) {
	tableName, err := getRequiredString(data, "tableName", filePath, changesetID, "dropDefaultValue")
	if err != nil {
		return Change{}, err
	}

	columnName, err := getRequiredString(data, "columnName", filePath, changesetID, "dropDefaultValue")
	if err != nil {
		return Change{}, err
	}

	return Change{
		Type:       "dropDefaultValue",
		TableName:  tableName,
		ColumnName: columnName,
		Attributes: flattenMap(data),
	}, nil
}

// convertCreateIndex converts a createIndex change
func convertCreateIndex(data map[string]any, filePath, changesetID string) (Change, error) {
	tableName, err := getRequiredString(data, "tableName", filePath, changesetID, "createIndex")
	if err != nil {
		return Change{}, err
	}

	indexName := getOptionalString(data, "indexName", "")

	return Change{
		Type:       "createIndex",
		TableName:  tableName,
		IndexName:  indexName,
		Attributes: flattenMap(data),
	}, nil
}

// convertDropIndex converts a dropIndex change
func convertDropIndex(data map[string]any, filePath, changesetID string) (Change, error) {
	tableName := getOptionalString(data, "tableName", "")
	indexName := getOptionalString(data, "indexName", "")

	// Either tableName or indexName is required
	if tableName == "" && indexName == "" {
		return Change{}, fmt.Errorf("%s:changeSet[%s]:dropIndex: missing required field 'tableName' or 'indexName'", filePath, changesetID)
	}

	return Change{
		Type:       "dropIndex",
		TableName:  tableName,
		IndexName:  indexName,
		Attributes: flattenMap(data),
	}, nil
}

// convertInsert converts an insert change
func convertInsert(data map[string]any, filePath, changesetID string) (Change, error) {
	tableName, err := getRequiredString(data, "tableName", filePath, changesetID, "insert")
	if err != nil {
		return Change{}, err
	}

	return Change{
		Type:       "insert",
		TableName:  tableName,
		Attributes: flattenMap(data),
	}, nil
}

// convertUpdate converts an update change
func convertUpdate(data map[string]any, filePath, changesetID string) (Change, error) {
	tableName, err := getRequiredString(data, "tableName", filePath, changesetID, "update")
	if err != nil {
		return Change{}, err
	}

	return Change{
		Type:       "update",
		TableName:  tableName,
		Attributes: flattenMap(data),
	}, nil
}

// convertDelete converts a delete change
func convertDelete(data map[string]any, filePath, changesetID string) (Change, error) {
	tableName, err := getRequiredString(data, "tableName", filePath, changesetID, "delete")
	if err != nil {
		return Change{}, err
	}

	return Change{
		Type:       "delete",
		TableName:  tableName,
		Attributes: flattenMap(data),
	}, nil
}

// convertSQL converts a sql or sqlFile change
func convertSQL(data map[string]any, _ /* filePath */, _ /* changesetID */, changeType string) (Change, error) {
	var sql string

	if changeType == "sql" {
		// For sql change, the SQL can be in the content or as a text node
		if sqlVal, ok := data["sql"].(string); ok {
			sql = sqlVal
		} else if content, ok := data["content"].(string); ok {
			sql = content
		} else if text, ok := data[""].(string); ok {
			// YAML/JSON may have text content with empty key
			sql = text
		}
	}

	return Change{
		Type:       changeType,
		SQL:        sql,
		Attributes: flattenMap(data),
	}, nil
}

// getRequiredString extracts a required string field from a map
func getRequiredString(data map[string]any, field, filePath, changesetID, changeType string) (string, error) {
	val, ok := data[field]
	if !ok {
		return "", fmt.Errorf("%s:changeSet[%s]:%s: missing required field '%s'", filePath, changesetID, changeType, field)
	}

	strVal, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("%s:changeSet[%s]:%s: field '%s' must be a string, got %T", filePath, changesetID, changeType, field, val)
	}

	if strVal == "" {
		return "", fmt.Errorf("%s:changeSet[%s]:%s: field '%s' cannot be empty", filePath, changesetID, changeType, field)
	}

	return strVal, nil
}

// getOptionalString extracts an optional string field from a map
func getOptionalString(data map[string]any, field, defaultVal string) string {
	val, ok := data[field]
	if !ok {
		return defaultVal
	}

	strVal, ok := val.(string)
	if !ok {
		return defaultVal
	}

	return strVal
}

// flattenMap converts a map[string]any to map[string]string for simple attributes
func flattenMap(data map[string]any) map[string]string {
	result := make(map[string]string)
	for k, v := range data {
		// Skip complex nested structures
		switch val := v.(type) {
		case string:
			result[k] = val
		case int, int64, int32:
			result[k] = fmt.Sprintf("%d", val)
		case float64, float32:
			result[k] = fmt.Sprintf("%f", val)
		case bool:
			result[k] = fmt.Sprintf("%t", val)
		case nil:
			// Skip nil values
		default:
			// For complex types, just note their presence
			result[k] = fmt.Sprintf("<%T>", val)
		}
	}
	return result
}
