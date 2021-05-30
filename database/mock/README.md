Factory package is a simply two function for each model that is used for mocking:

- `func MakeModel(in Model) (out Model, e error)` - creates instance of `Model` with random field values or from passed in `Model`
- `func CreateModel(in Model) (out Model, e error)` - calls `MakeModel` and inserts result into database

The functions body is up to user and situation. See `topic.go` as basic example and `entity.go` as advanced 