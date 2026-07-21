implement as follows:

- wherever you declared []interface{} typically for passing parameters to a parametrized database query, replace with []any for cosmetic
- wherever the value 'owner' is present in SQL queries, please pass it as bound parameter, and use for that purpose the constant defined in @app.go
