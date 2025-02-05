package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/fraenky8/tables-to-go/pkg/database"
	"github.com/fraenky8/tables-to-go/pkg/settings"
)

type mockDb struct {
	mock.Mock
	database.Database

	tables []*database.Table
}

func newMockDb(db database.Database) *mockDb {
	return &mockDb{Database: db}
}

func (db *mockDb) Connect() (err error) {
	db.Called()
	return nil
}

func (db *mockDb) Close() (err error) {
	db.Called()
	return nil
}

func (db *mockDb) GetTables() (tables []*database.Table, err error) {
	db.Called()
	return db.tables, nil
}

func (db *mockDb) PrepareGetColumnsOfTableStmt() (err error) {
	db.Called()
	return nil
}

func (db *mockDb) GetColumnsOfTable(table *database.Table) (err error) {
	db.Called(table)
	return nil
}

type mockWriter struct {
	mock.Mock
}

func newMockWriter() *mockWriter {
	return &mockWriter{}
}

func (w *mockWriter) Write(tableName string, content string) error {
	w.Called(tableName, content)
	return nil
}

func TestCamelCaseString(t *testing.T) {
	tests := []struct {
		desc     string
		input    string
		expected string
	}{
		{
			desc:     "empty string returns empty string",
			input:    "",
			expected: "",
		},
		{
			desc:     "single string returns titleized single string",
			input:    "string",
			expected: "String",
		},
		{
			desc:     "multi separated string returns CamelCase string",
			input:    "string_with_separate_sections",
			expected: "StringWithSeparateSections",
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			actual := camelCaseString(tt.input)
			assert.Equal(t, tt.expected, actual, "test case input: "+tt.input)
		})
	}
}

func TestToInitialisms(t *testing.T) {
	tests := []struct {
		desc     string
		input    string
		expected string
	}{
		{
			desc:     "id should be upper case",
			input:    "Id",
			expected: "ID",
		},
		{
			desc:     "id at the end of string should be upper case",
			input:    "userId",
			expected: "userID",
		},
		{
			desc:     "id at the beginning of string should be upper case",
			input:    "Iduser",
			expected: "IDuser",
		},
		{
			desc:     "id in the middle of string should be upper case",
			input:    "userIdprim",
			expected: "userIDprim",
		},
		{
			desc:     "multiple occurrences should be upper case",
			input:    "userIdasJsonWithUrl",
			expected: "userIDasJSONWithURL",
		},
		{
			desc:     "multiple id in the string should be upper case",
			input:    "IduserId",
			expected: "IDuserID",
		},
		{
			desc:     "non replacement in the string should be return original string",
			input:    "name",
			expected: "name",
		},
		{
			desc:     "replacements only in the string should be return original string",
			input:    "IdjsonuRlHtTp",
			expected: "IDJSONURLHTTP",
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			actual := toInitialisms(tt.input)
			assert.Equal(t, tt.expected, actual, "test case input: "+tt.input)
		})
	}
}

func TestRun_StringTextColumns(t *testing.T) {
	for dbType := range settings.SupportedDbTypes {
		t.Run(dbType.String(), func(t *testing.T) {

			s := settings.New()
			s.DbType = dbType
			db := database.New(s)

			columnTypes := db.GetStringDatatypes()

			for _, columnType := range columnTypes {
				t.Run(columnType, func(t *testing.T) {

					t.Run("single table with NOT NULL column", func(t *testing.T) {
						s := settings.New()
						s.DbType = dbType

						mdb := newMockDb(db)

						table := &database.Table{
							Name: "test_table",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name",
									DataType:        columnType,
								},
							},
						}
						mdb.tables = append(mdb.tables, table)

						mdb.
							On("GetTables").
							Return(mdb.tables, nil)
						mdb.
							On("PrepareGetColumnsOfTableStmt").
							Return(nil)
						mdb.
							On("GetColumnsOfTable", table)

						w := newMockWriter()
						w.
							On(
								"Write",
								"TestTable",
								"package dto\n\ntype TestTable struct {\nColumnName string `db:\"column_name\"`\n}",
							)

						err := Run(s, mdb, w)
						assert.NoError(t, err)
					})

					t.Run("single table with NULL column", func(t *testing.T) {
						s := settings.New()
						s.DbType = dbType

						mdb := newMockDb(db)

						table := &database.Table{
							Name: "test_table",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name",
									DataType:        columnType,
									IsNullable:      "YES",
								},
							},
						}
						mdb.tables = append(mdb.tables, table)

						mdb.
							On("GetTables").
							Return(mdb.tables, nil)
						mdb.
							On("PrepareGetColumnsOfTableStmt").
							Return(nil)
						mdb.
							On("GetColumnsOfTable", table)

						w := newMockWriter()
						w.
							On(
								"Write",
								"TestTable",
								"package dto\n\nimport (\n\t\"database/sql\"\n)\n\ntype TestTable struct {\nColumnName sql.NullString `db:\"column_name\"`\n}",
							)

						err := Run(s, mdb, w)
						assert.NoError(t, err)
					})

					t.Run("single table with NULL column and native data type", func(t *testing.T) {
						s := settings.New()
						s.DbType = dbType
						s.Null = settings.NullTypeNative

						mdb := newMockDb(db)

						table := &database.Table{
							Name: "test_table",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name",
									DataType:        columnType,
									IsNullable:      "YES",
								},
							},
						}
						mdb.tables = append(mdb.tables, table)

						mdb.
							On("GetTables").
							Return(mdb.tables, nil)
						mdb.
							On("PrepareGetColumnsOfTableStmt").
							Return(nil)
						mdb.
							On("GetColumnsOfTable", table)

						w := newMockWriter()
						w.
							On(
								"Write",
								"TestTable",
								"package dto\n\nimport (\n)\n\ntype TestTable struct {\nColumnName *string `db:\"column_name\"`\n}",
							)

						err := Run(s, mdb, w)
						assert.NoError(t, err)
					})

					t.Run("single table with two mixed columns", func(t *testing.T) {
						s := settings.New()
						s.DbType = dbType

						mdb := newMockDb(db)

						table := &database.Table{
							Name: "test_table",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name_1",
									DataType:        columnType,
									IsNullable:      "YES",
								},
								{
									OrdinalPosition: 2,
									Name:            "column_name_2",
									DataType:        columnType,
								},
							},
						}
						mdb.tables = append(mdb.tables, table)

						mdb.
							On("GetTables").
							Return(mdb.tables, nil)
						mdb.
							On("PrepareGetColumnsOfTableStmt").
							Return(nil)
						mdb.
							On("GetColumnsOfTable", table)

						w := newMockWriter()
						w.
							On(
								"Write",
								"TestTable",
								"package dto\n\nimport (\n\t\"database/sql\"\n)\n\ntype TestTable struct {\nColumnName1 sql.NullString `db:\"column_name_1\"`\nColumnName2 string `db:\"column_name_2\"`\n}",
							)

						err := Run(s, mdb, w)
						assert.NoError(t, err)
					})

					t.Run("single table with two mixed columns and native data type", func(t *testing.T) {
						s := settings.New()
						s.DbType = dbType
						s.Null = settings.NullTypeNative

						mdb := newMockDb(db)

						table := &database.Table{
							Name: "test_table",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name_1",
									DataType:        columnType,
									IsNullable:      "YES",
								},
								{
									OrdinalPosition: 2,
									Name:            "column_name_2",
									DataType:        columnType,
								},
							},
						}
						mdb.tables = append(mdb.tables, table)

						mdb.
							On("GetTables").
							Return(mdb.tables, nil)
						mdb.
							On("PrepareGetColumnsOfTableStmt").
							Return(nil)
						mdb.
							On("GetColumnsOfTable", table)

						w := newMockWriter()
						w.
							On(
								"Write",
								"TestTable",
								"package dto\n\nimport (\n)\n\ntype TestTable struct {\nColumnName1 *string `db:\"column_name_1\"`\nColumnName2 string `db:\"column_name_2\"`\n}",
							)

						err := Run(s, mdb, w)
						assert.NoError(t, err)
					})

					t.Run("multi table with multi columns", func(t *testing.T) {
						s := settings.New()
						s.DbType = dbType

						mdb := newMockDb(db)

						table1 := &database.Table{
							Name: "test_table_1",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name_1",
									DataType:        columnType,
									IsNullable:      "YES",
								},
								{
									OrdinalPosition: 2,
									Name:            "column_name_2",
									DataType:        columnType,
								},
							},
						}
						table2 := &database.Table{
							Name: "test_table_2",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name_1",
									DataType:        columnType,
								},
								{
									OrdinalPosition: 2,
									Name:            "column_name_2",
									DataType:        columnType,
									IsNullable:      "YES",
								},
							},
						}
						mdb.tables = append(mdb.tables, table1, table2)

						mdb.
							On("GetTables").
							Return(mdb.tables, nil)
						mdb.
							On("PrepareGetColumnsOfTableStmt").
							Return(nil)
						mdb.
							On("GetColumnsOfTable", table1).
							On("GetColumnsOfTable", table2)

						w := newMockWriter()
						w.
							On(
								"Write",
								"TestTable1",
								"package dto\n\nimport (\n\t\"database/sql\"\n)\n\ntype TestTable1 struct {\nColumnName1 sql.NullString `db:\"column_name_1\"`\nColumnName2 string `db:\"column_name_2\"`\n}",
							).
							On(
								"Write",
								"TestTable2",
								"package dto\n\nimport (\n\t\"database/sql\"\n)\n\ntype TestTable2 struct {\nColumnName1 string `db:\"column_name_1\"`\nColumnName2 sql.NullString `db:\"column_name_2\"`\n}",
							)

						err := Run(s, mdb, w)
						assert.NoError(t, err)
					})
				})
			}
		})
	}
}

func TestRun_IntegerColumns(t *testing.T) {
	for dbType := range settings.SupportedDbTypes {
		t.Run(dbType.String(), func(t *testing.T) {

			s := settings.New()
			s.DbType = dbType
			db := database.New(s)

			columnTypes := db.GetIntegerDatatypes()

			for _, columnType := range columnTypes {
				t.Run(columnType, func(t *testing.T) {

					t.Run("single table with NOT NULL column", func(t *testing.T) {
						s := settings.New()
						s.DbType = dbType

						mdb := newMockDb(db)

						table := &database.Table{
							Name: "test_table",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name",
									DataType:        columnType,
								},
							},
						}
						mdb.tables = append(mdb.tables, table)

						mdb.
							On("GetTables").
							Return(mdb.tables, nil)
						mdb.
							On("PrepareGetColumnsOfTableStmt").
							Return(nil)
						mdb.
							On("GetColumnsOfTable", table)

						w := newMockWriter()
						w.
							On(
								"Write",
								"TestTable",
								"package dto\n\ntype TestTable struct {\nColumnName int `db:\"column_name\"`\n}",
							)

						err := Run(s, mdb, w)
						assert.NoError(t, err)
					})

					t.Run("single table with NULL column", func(t *testing.T) {
						s := settings.New()
						s.DbType = dbType

						mdb := newMockDb(db)

						table := &database.Table{
							Name: "test_table",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name",
									DataType:        columnType,
									IsNullable:      "YES",
								},
							},
						}
						mdb.tables = append(mdb.tables, table)

						mdb.
							On("GetTables").
							Return(mdb.tables, nil)
						mdb.
							On("PrepareGetColumnsOfTableStmt").
							Return(nil)
						mdb.
							On("GetColumnsOfTable", table)

						w := newMockWriter()
						w.
							On(
								"Write",
								"TestTable",
								"package dto\n\nimport (\n\t\"database/sql\"\n)\n\ntype TestTable struct {\nColumnName sql.NullInt64 `db:\"column_name\"`\n}",
							)

						err := Run(s, mdb, w)
						assert.NoError(t, err)
					})

					t.Run("single table with NULL column and native data type", func(t *testing.T) {
						s := settings.New()
						s.DbType = dbType
						s.Null = settings.NullTypeNative

						mdb := newMockDb(db)

						table := &database.Table{
							Name: "test_table",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name",
									DataType:        columnType,
									IsNullable:      "YES",
								},
							},
						}
						mdb.tables = append(mdb.tables, table)

						mdb.
							On("GetTables").
							Return(mdb.tables, nil)
						mdb.
							On("PrepareGetColumnsOfTableStmt").
							Return(nil)
						mdb.
							On("GetColumnsOfTable", table)

						w := newMockWriter()
						w.
							On(
								"Write",
								"TestTable",
								"package dto\n\nimport (\n)\n\ntype TestTable struct {\nColumnName *int `db:\"column_name\"`\n}",
							)

						err := Run(s, mdb, w)
						assert.NoError(t, err)
					})

					t.Run("single table with two mixed columns", func(t *testing.T) {
						s := settings.New()
						s.DbType = dbType

						mdb := newMockDb(db)

						table := &database.Table{
							Name: "test_table",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name_1",
									DataType:        columnType,
									IsNullable:      "YES",
								},
								{
									OrdinalPosition: 2,
									Name:            "column_name_2",
									DataType:        columnType,
								},
							},
						}
						mdb.tables = append(mdb.tables, table)

						mdb.
							On("GetTables").
							Return(mdb.tables, nil)
						mdb.
							On("PrepareGetColumnsOfTableStmt").
							Return(nil)
						mdb.
							On("GetColumnsOfTable", table)

						w := newMockWriter()
						w.
							On(
								"Write",
								"TestTable",
								"package dto\n\nimport (\n\t\"database/sql\"\n)\n\ntype TestTable struct {\nColumnName1 sql.NullInt64 `db:\"column_name_1\"`\nColumnName2 int `db:\"column_name_2\"`\n}",
							)

						err := Run(s, mdb, w)
						assert.NoError(t, err)
					})

					t.Run("single table with two mixed columns and native data type", func(t *testing.T) {
						s := settings.New()
						s.DbType = dbType
						s.Null = settings.NullTypeNative

						mdb := newMockDb(db)

						table := &database.Table{
							Name: "test_table",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name_1",
									DataType:        columnType,
									IsNullable:      "YES",
								},
								{
									OrdinalPosition: 2,
									Name:            "column_name_2",
									DataType:        columnType,
								},
							},
						}
						mdb.tables = append(mdb.tables, table)

						mdb.
							On("GetTables").
							Return(mdb.tables, nil)
						mdb.
							On("PrepareGetColumnsOfTableStmt").
							Return(nil)
						mdb.
							On("GetColumnsOfTable", table)

						w := newMockWriter()
						w.
							On(
								"Write",
								"TestTable",
								"package dto\n\nimport (\n)\n\ntype TestTable struct {\nColumnName1 *int `db:\"column_name_1\"`\nColumnName2 int `db:\"column_name_2\"`\n}",
							)

						err := Run(s, mdb, w)
						assert.NoError(t, err)
					})

					t.Run("multi table with multi columns", func(t *testing.T) {
						s := settings.New()
						s.DbType = dbType

						mdb := newMockDb(db)

						table1 := &database.Table{
							Name: "test_table_1",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name_1",
									DataType:        columnType,
									IsNullable:      "YES",
								},
								{
									OrdinalPosition: 2,
									Name:            "column_name_2",
									DataType:        columnType,
								},
							},
						}
						table2 := &database.Table{
							Name: "test_table_2",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name_1",
									DataType:        columnType,
								},
								{
									OrdinalPosition: 2,
									Name:            "column_name_2",
									DataType:        columnType,
									IsNullable:      "YES",
								},
							},
						}
						mdb.tables = append(mdb.tables, table1, table2)

						mdb.
							On("GetTables").
							Return(mdb.tables, nil)
						mdb.
							On("PrepareGetColumnsOfTableStmt").
							Return(nil)
						mdb.
							On("GetColumnsOfTable", table1).
							On("GetColumnsOfTable", table2)

						w := newMockWriter()
						w.
							On(
								"Write",
								"TestTable1",
								"package dto\n\nimport (\n\t\"database/sql\"\n)\n\ntype TestTable1 struct {\nColumnName1 sql.NullInt64 `db:\"column_name_1\"`\nColumnName2 int `db:\"column_name_2\"`\n}",
							).
							On(
								"Write",
								"TestTable2",
								"package dto\n\nimport (\n\t\"database/sql\"\n)\n\ntype TestTable2 struct {\nColumnName1 int `db:\"column_name_1\"`\nColumnName2 sql.NullInt64 `db:\"column_name_2\"`\n}",
							)

						err := Run(s, mdb, w)
						assert.NoError(t, err)
					})
				})
			}
		})
	}
}

func TestRun_FloatColumns(t *testing.T) {
	for dbType := range settings.SupportedDbTypes {
		t.Run(dbType.String(), func(t *testing.T) {

			s := settings.New()
			s.DbType = dbType
			db := database.New(s)

			columnTypes := db.GetFloatDatatypes()

			for _, columnType := range columnTypes {
				t.Run(columnType, func(t *testing.T) {

					t.Run("single table with NOT NULL column", func(t *testing.T) {
						s := settings.New()
						s.DbType = dbType

						mdb := newMockDb(db)

						table := &database.Table{
							Name: "test_table",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name",
									DataType:        columnType,
								},
							},
						}
						mdb.tables = append(mdb.tables, table)

						mdb.
							On("GetTables").
							Return(mdb.tables, nil)
						mdb.
							On("PrepareGetColumnsOfTableStmt").
							Return(nil)
						mdb.
							On("GetColumnsOfTable", table)

						w := newMockWriter()
						w.
							On(
								"Write",
								"TestTable",
								"package dto\n\ntype TestTable struct {\nColumnName float64 `db:\"column_name\"`\n}",
							)

						err := Run(s, mdb, w)
						assert.NoError(t, err)
					})

					t.Run("single table with NULL column", func(t *testing.T) {
						s := settings.New()
						s.DbType = dbType

						mdb := newMockDb(db)

						table := &database.Table{
							Name: "test_table",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name",
									DataType:        columnType,
									IsNullable:      "YES",
								},
							},
						}
						mdb.tables = append(mdb.tables, table)

						mdb.
							On("GetTables").
							Return(mdb.tables, nil)
						mdb.
							On("PrepareGetColumnsOfTableStmt").
							Return(nil)
						mdb.
							On("GetColumnsOfTable", table)

						w := newMockWriter()
						w.
							On(
								"Write",
								"TestTable",
								"package dto\n\nimport (\n\t\"database/sql\"\n)\n\ntype TestTable struct {\nColumnName sql.NullFloat64 `db:\"column_name\"`\n}",
							)

						err := Run(s, mdb, w)
						assert.NoError(t, err)
					})

					t.Run("single table with NULL column and native data type", func(t *testing.T) {
						s := settings.New()
						s.DbType = dbType
						s.Null = settings.NullTypeNative

						mdb := newMockDb(db)

						table := &database.Table{
							Name: "test_table",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name",
									DataType:        columnType,
									IsNullable:      "YES",
								},
							},
						}
						mdb.tables = append(mdb.tables, table)

						mdb.
							On("GetTables").
							Return(mdb.tables, nil)
						mdb.
							On("PrepareGetColumnsOfTableStmt").
							Return(nil)
						mdb.
							On("GetColumnsOfTable", table)

						w := newMockWriter()
						w.
							On(
								"Write",
								"TestTable",
								"package dto\n\nimport (\n)\n\ntype TestTable struct {\nColumnName *float64 `db:\"column_name\"`\n}",
							)

						err := Run(s, mdb, w)
						assert.NoError(t, err)
					})

					t.Run("single table with two mixed columns", func(t *testing.T) {
						s := settings.New()
						s.DbType = dbType

						mdb := newMockDb(db)

						table := &database.Table{
							Name: "test_table",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name_1",
									DataType:        columnType,
									IsNullable:      "YES",
								},
								{
									OrdinalPosition: 2,
									Name:            "column_name_2",
									DataType:        columnType,
								},
							},
						}
						mdb.tables = append(mdb.tables, table)

						mdb.
							On("GetTables").
							Return(mdb.tables, nil)
						mdb.
							On("PrepareGetColumnsOfTableStmt").
							Return(nil)
						mdb.
							On("GetColumnsOfTable", table)

						w := newMockWriter()
						w.
							On(
								"Write",
								"TestTable",
								"package dto\n\nimport (\n\t\"database/sql\"\n)\n\ntype TestTable struct {\nColumnName1 sql.NullFloat64 `db:\"column_name_1\"`\nColumnName2 float64 `db:\"column_name_2\"`\n}",
							)

						err := Run(s, mdb, w)
						assert.NoError(t, err)
					})

					t.Run("single table with two mixed columns and native data type", func(t *testing.T) {
						s := settings.New()
						s.DbType = dbType
						s.Null = settings.NullTypeNative

						mdb := newMockDb(db)

						table := &database.Table{
							Name: "test_table",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name_1",
									DataType:        columnType,
									IsNullable:      "YES",
								},
								{
									OrdinalPosition: 2,
									Name:            "column_name_2",
									DataType:        columnType,
								},
							},
						}
						mdb.tables = append(mdb.tables, table)

						mdb.
							On("GetTables").
							Return(mdb.tables, nil)
						mdb.
							On("PrepareGetColumnsOfTableStmt").
							Return(nil)
						mdb.
							On("GetColumnsOfTable", table)

						w := newMockWriter()
						w.
							On(
								"Write",
								"TestTable",
								"package dto\n\nimport (\n)\n\ntype TestTable struct {\nColumnName1 *float64 `db:\"column_name_1\"`\nColumnName2 float64 `db:\"column_name_2\"`\n}",
							)

						err := Run(s, mdb, w)
						assert.NoError(t, err)
					})

					t.Run("multi table with multi columns", func(t *testing.T) {
						s := settings.New()
						s.DbType = dbType

						mdb := newMockDb(db)

						table1 := &database.Table{
							Name: "test_table_1",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name_1",
									DataType:        columnType,
									IsNullable:      "YES",
								},
								{
									OrdinalPosition: 2,
									Name:            "column_name_2",
									DataType:        columnType,
								},
							},
						}
						table2 := &database.Table{
							Name: "test_table_2",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name_1",
									DataType:        columnType,
								},
								{
									OrdinalPosition: 2,
									Name:            "column_name_2",
									DataType:        columnType,
									IsNullable:      "YES",
								},
							},
						}
						mdb.tables = append(mdb.tables, table1, table2)

						mdb.
							On("GetTables").
							Return(mdb.tables, nil)
						mdb.
							On("PrepareGetColumnsOfTableStmt").
							Return(nil)
						mdb.
							On("GetColumnsOfTable", table1).
							On("GetColumnsOfTable", table2)

						w := newMockWriter()
						w.
							On(
								"Write",
								"TestTable1",
								"package dto\n\nimport (\n\t\"database/sql\"\n)\n\ntype TestTable1 struct {\nColumnName1 sql.NullFloat64 `db:\"column_name_1\"`\nColumnName2 float64 `db:\"column_name_2\"`\n}",
							).
							On(
								"Write",
								"TestTable2",
								"package dto\n\nimport (\n\t\"database/sql\"\n)\n\ntype TestTable2 struct {\nColumnName1 float64 `db:\"column_name_1\"`\nColumnName2 sql.NullFloat64 `db:\"column_name_2\"`\n}",
							)

						err := Run(s, mdb, w)
						assert.NoError(t, err)
					})
				})
			}
		})
	}
}

func TestRun_TemporalColumns(t *testing.T) {
	for dbType := range settings.SupportedDbTypes {
		t.Run(dbType.String(), func(t *testing.T) {

			s := settings.New()
			s.DbType = dbType
			db := database.New(s)

			columnTypes := db.GetTemporalDatatypes()

			for _, columnType := range columnTypes {
				t.Run(columnType, func(t *testing.T) {

					t.Run("single table with NOT NULL column", func(t *testing.T) {
						s := settings.New()
						s.DbType = dbType

						mdb := newMockDb(db)

						table := &database.Table{
							Name: "test_table",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name",
									DataType:        columnType,
								},
							},
						}
						mdb.tables = append(mdb.tables, table)

						mdb.
							On("GetTables").
							Return(mdb.tables, nil)
						mdb.
							On("PrepareGetColumnsOfTableStmt").
							Return(nil)
						mdb.
							On("GetColumnsOfTable", table)

						w := newMockWriter()
						w.
							On(
								"Write",
								"TestTable",
								"package dto\n\nimport (\n\t\"time\"\n)\n\ntype TestTable struct {\nColumnName time.Time `db:\"column_name\"`\n}",
							)

						err := Run(s, mdb, w)
						assert.NoError(t, err)
					})

					t.Run("single table with NULL column", func(t *testing.T) {
						s := settings.New()
						s.DbType = dbType

						mdb := newMockDb(db)

						table := &database.Table{
							Name: "test_table",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name",
									DataType:        columnType,
									IsNullable:      "YES",
								},
							},
						}
						mdb.tables = append(mdb.tables, table)

						mdb.
							On("GetTables").
							Return(mdb.tables, nil)
						mdb.
							On("PrepareGetColumnsOfTableStmt").
							Return(nil)
						mdb.
							On("GetColumnsOfTable", table)

						w := newMockWriter()
						w.
							On(
								"Write",
								"TestTable",
								"package dto\n\nimport (\n\t\n"+db.GetDriverImportLibrary()+"\n)\n\ntype TestTable struct {\nColumnName "+dbType.String()+".NullTime `db:\"column_name\"`\n}",
							)

						err := Run(s, mdb, w)
						assert.NoError(t, err)
					})

					t.Run("single table with NULL column and native data type", func(t *testing.T) {
						s := settings.New()
						s.DbType = dbType
						s.Null = settings.NullTypeNative

						mdb := newMockDb(db)

						table := &database.Table{
							Name: "test_table",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name",
									DataType:        columnType,
									IsNullable:      "YES",
								},
							},
						}
						mdb.tables = append(mdb.tables, table)

						mdb.
							On("GetTables").
							Return(mdb.tables, nil)
						mdb.
							On("PrepareGetColumnsOfTableStmt").
							Return(nil)
						mdb.
							On("GetColumnsOfTable", table)

						w := newMockWriter()
						w.
							On(
								"Write",
								"TestTable",
								"package dto\n\nimport (\n\t\"time\"\n)\n\ntype TestTable struct {\nColumnName *time.Time `db:\"column_name\"`\n}",
							)

						err := Run(s, mdb, w)
						assert.NoError(t, err)
					})

					t.Run("single table with two mixed columns", func(t *testing.T) {
						s := settings.New()
						s.DbType = dbType

						mdb := newMockDb(db)

						table := &database.Table{
							Name: "test_table",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name_1",
									DataType:        columnType,
									IsNullable:      "YES",
								},
								{
									OrdinalPosition: 2,
									Name:            "column_name_2",
									DataType:        columnType,
								},
							},
						}
						mdb.tables = append(mdb.tables, table)

						mdb.
							On("GetTables").
							Return(mdb.tables, nil)
						mdb.
							On("PrepareGetColumnsOfTableStmt").
							Return(nil)
						mdb.
							On("GetColumnsOfTable", table)

						w := newMockWriter()
						w.
							On(
								"Write",
								"TestTable",
								"package dto\n\nimport (\n\t\"time\"\n\t\n"+db.GetDriverImportLibrary()+"\n)\n\ntype TestTable struct {\nColumnName1 "+dbType.String()+".NullTime `db:\"column_name_1\"`\nColumnName2 time.Time `db:\"column_name_2\"`\n}",
							)

						err := Run(s, mdb, w)
						assert.NoError(t, err)
					})

					t.Run("single table with two mixed columns and native data type", func(t *testing.T) {
						s := settings.New()
						s.DbType = dbType
						s.Null = settings.NullTypeNative

						mdb := newMockDb(db)

						table := &database.Table{
							Name: "test_table",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name_1",
									DataType:        columnType,
									IsNullable:      "YES",
								},
								{
									OrdinalPosition: 2,
									Name:            "column_name_2",
									DataType:        columnType,
								},
							},
						}
						mdb.tables = append(mdb.tables, table)

						mdb.
							On("GetTables").
							Return(mdb.tables, nil)
						mdb.
							On("PrepareGetColumnsOfTableStmt").
							Return(nil)
						mdb.
							On("GetColumnsOfTable", table)

						w := newMockWriter()
						w.
							On(
								"Write",
								"TestTable",
								"package dto\n\nimport (\n\t\"time\"\n)\n\ntype TestTable struct {\nColumnName1 *time.Time `db:\"column_name_1\"`\nColumnName2 time.Time `db:\"column_name_2\"`\n}",
							)

						err := Run(s, mdb, w)
						assert.NoError(t, err)
					})

					t.Run("multi table with multi columns", func(t *testing.T) {
						s := settings.New()
						s.DbType = dbType

						mdb := newMockDb(db)

						table1 := &database.Table{
							Name: "test_table_1",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name_1",
									DataType:        columnType,
									IsNullable:      "YES",
								},
								{
									OrdinalPosition: 2,
									Name:            "column_name_2",
									DataType:        columnType,
								},
							},
						}
						table2 := &database.Table{
							Name: "test_table_2",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name_1",
									DataType:        columnType,
								},
								{
									OrdinalPosition: 2,
									Name:            "column_name_2",
									DataType:        columnType,
									IsNullable:      "YES",
								},
							},
						}
						mdb.tables = append(mdb.tables, table1, table2)

						mdb.
							On("GetTables").
							Return(mdb.tables, nil)
						mdb.
							On("PrepareGetColumnsOfTableStmt").
							Return(nil)
						mdb.
							On("GetColumnsOfTable", table1).
							On("GetColumnsOfTable", table2)

						w := newMockWriter()
						w.
							On(
								"Write",
								"TestTable1",
								"package dto\n\nimport (\n\t\"time\"\n\t\n"+db.GetDriverImportLibrary()+"\n)\n\ntype TestTable1 struct {\nColumnName1 "+dbType.String()+".NullTime `db:\"column_name_1\"`\nColumnName2 time.Time `db:\"column_name_2\"`\n}",
							).
							On(
								"Write",
								"TestTable2",
								"package dto\n\nimport (\n\t\"time\"\n\t\n"+db.GetDriverImportLibrary()+"\n)\n\ntype TestTable2 struct {\nColumnName1 time.Time `db:\"column_name_1\"`\nColumnName2 "+dbType.String()+".NullTime `db:\"column_name_2\"`\n}",
							)

						err := Run(s, mdb, w)
						assert.NoError(t, err)
					})
				})
			}
		})
	}
}

func TestRun_BooleanColumns(t *testing.T) {
	for dbType := range settings.SupportedDbTypes {
		t.Run(dbType.String(), func(t *testing.T) {

			s := settings.New()
			s.DbType = dbType
			db := database.New(s)

			columnTypes := []string{"boolean"}

			for _, columnType := range columnTypes {
				t.Run(columnType, func(t *testing.T) {

					t.Run("single table with NOT NULL column", func(t *testing.T) {
						s := settings.New()
						s.DbType = dbType

						mdb := newMockDb(db)

						table := &database.Table{
							Name: "test_table",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name",
									DataType:        columnType,
								},
							},
						}
						mdb.tables = append(mdb.tables, table)

						mdb.
							On("GetTables").
							Return(mdb.tables, nil)
						mdb.
							On("PrepareGetColumnsOfTableStmt").
							Return(nil)
						mdb.
							On("GetColumnsOfTable", table)

						w := newMockWriter()
						w.
							On(
								"Write",
								"TestTable",
								"package dto\n\ntype TestTable struct {\nColumnName bool `db:\"column_name\"`\n}",
							)

						err := Run(s, mdb, w)
						assert.NoError(t, err)
					})

					t.Run("single table with NULL column", func(t *testing.T) {
						s := settings.New()
						s.DbType = dbType

						mdb := newMockDb(db)

						table := &database.Table{
							Name: "test_table",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name",
									DataType:        columnType,
									IsNullable:      "YES",
								},
							},
						}
						mdb.tables = append(mdb.tables, table)

						mdb.
							On("GetTables").
							Return(mdb.tables, nil)
						mdb.
							On("PrepareGetColumnsOfTableStmt").
							Return(nil)
						mdb.
							On("GetColumnsOfTable", table)

						w := newMockWriter()
						w.
							On(
								"Write",
								"TestTable",
								"package dto\n\nimport (\n\t\"database/sql\"\n)\n\ntype TestTable struct {\nColumnName sql.NullBool `db:\"column_name\"`\n}",
							)

						err := Run(s, mdb, w)
						assert.NoError(t, err)
					})

					t.Run("single table with NULL column and native data type", func(t *testing.T) {
						s := settings.New()
						s.DbType = dbType
						s.Null = settings.NullTypeNative

						mdb := newMockDb(db)

						table := &database.Table{
							Name: "test_table",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name",
									DataType:        columnType,
									IsNullable:      "YES",
								},
							},
						}
						mdb.tables = append(mdb.tables, table)

						mdb.
							On("GetTables").
							Return(mdb.tables, nil)
						mdb.
							On("PrepareGetColumnsOfTableStmt").
							Return(nil)
						mdb.
							On("GetColumnsOfTable", table)

						w := newMockWriter()
						w.
							On(
								"Write",
								"TestTable",
								"package dto\n\nimport (\n)\n\ntype TestTable struct {\nColumnName *bool `db:\"column_name\"`\n}",
							)

						err := Run(s, mdb, w)
						assert.NoError(t, err)
					})

					t.Run("single table with two mixed columns", func(t *testing.T) {
						s := settings.New()
						s.DbType = dbType

						mdb := newMockDb(db)

						table := &database.Table{
							Name: "test_table",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name_1",
									DataType:        columnType,
									IsNullable:      "YES",
								},
								{
									OrdinalPosition: 2,
									Name:            "column_name_2",
									DataType:        columnType,
								},
							},
						}
						mdb.tables = append(mdb.tables, table)

						mdb.
							On("GetTables").
							Return(mdb.tables, nil)
						mdb.
							On("PrepareGetColumnsOfTableStmt").
							Return(nil)
						mdb.
							On("GetColumnsOfTable", table)

						w := newMockWriter()
						w.
							On(
								"Write",
								"TestTable",
								"package dto\n\nimport (\n\t\"database/sql\"\n)\n\ntype TestTable struct {\nColumnName1 sql.NullBool `db:\"column_name_1\"`\nColumnName2 bool `db:\"column_name_2\"`\n}",
							)

						err := Run(s, mdb, w)
						assert.NoError(t, err)
					})

					t.Run("single table with two mixed columns and native data type", func(t *testing.T) {
						s := settings.New()
						s.DbType = dbType
						s.Null = settings.NullTypeNative

						mdb := newMockDb(db)

						table := &database.Table{
							Name: "test_table",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name_1",
									DataType:        columnType,
									IsNullable:      "YES",
								},
								{
									OrdinalPosition: 2,
									Name:            "column_name_2",
									DataType:        columnType,
								},
							},
						}
						mdb.tables = append(mdb.tables, table)

						mdb.
							On("GetTables").
							Return(mdb.tables, nil)
						mdb.
							On("PrepareGetColumnsOfTableStmt").
							Return(nil)
						mdb.
							On("GetColumnsOfTable", table)

						w := newMockWriter()
						w.
							On(
								"Write",
								"TestTable",
								"package dto\n\nimport (\n)\n\ntype TestTable struct {\nColumnName1 *bool `db:\"column_name_1\"`\nColumnName2 bool `db:\"column_name_2\"`\n}",
							)

						err := Run(s, mdb, w)
						assert.NoError(t, err)
					})

					t.Run("multi table with multi columns", func(t *testing.T) {
						s := settings.New()
						s.DbType = dbType

						mdb := newMockDb(db)

						table1 := &database.Table{
							Name: "test_table_1",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name_1",
									DataType:        columnType,
									IsNullable:      "YES",
								},
								{
									OrdinalPosition: 2,
									Name:            "column_name_2",
									DataType:        columnType,
								},
							},
						}
						table2 := &database.Table{
							Name: "test_table_2",
							Columns: []database.Column{
								{
									OrdinalPosition: 1,
									Name:            "column_name_1",
									DataType:        columnType,
								},
								{
									OrdinalPosition: 2,
									Name:            "column_name_2",
									DataType:        columnType,
									IsNullable:      "YES",
								},
							},
						}
						mdb.tables = append(mdb.tables, table1, table2)

						mdb.
							On("GetTables").
							Return(mdb.tables, nil)
						mdb.
							On("PrepareGetColumnsOfTableStmt").
							Return(nil)
						mdb.
							On("GetColumnsOfTable", table1).
							On("GetColumnsOfTable", table2)

						w := newMockWriter()
						w.
							On(
								"Write",
								"TestTable1",
								"package dto\n\nimport (\n\t\"database/sql\"\n)\n\ntype TestTable1 struct {\nColumnName1 sql.NullBool `db:\"column_name_1\"`\nColumnName2 bool `db:\"column_name_2\"`\n}",
							).
							On(
								"Write",
								"TestTable2",
								"package dto\n\nimport (\n\t\"database/sql\"\n)\n\ntype TestTable2 struct {\nColumnName1 bool `db:\"column_name_1\"`\nColumnName2 sql.NullBool `db:\"column_name_2\"`\n}",
							)

						err := Run(s, mdb, w)
						assert.NoError(t, err)
					})
				})
			}
		})
	}
}
