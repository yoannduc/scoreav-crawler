# DynamoDB Data Modelisation

## Sample data and model

You can find in `doc/metal-news.nosql-workbench.json` a [nosql-workbench](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/workbench.settingup.html) export with (hopefully) up to date model and sample data

## Access Patterns

| #   | Access Patterns                            | Key Condition                                                     | Filter Condition                     |
| --- | ------------------------------------------ | :---------------------------------------------------------------- | ------------------------------------ |
| 1   | Get all by source and category             | pk = `<source>` and begins_with(sk, `<category>`)                 |                                      |
| 2   | Get all by source, category and date range | pk = `<source>`, sk=begins_with(`<category>`)                     | date between `<date1>` and `<date2>` |
| 3   | Get by id                                  | gsi_uuid:id = `<id>`                                              |                                      |
| 4   | Get uri list by day                        | gsi_date:date=`<date>`                                            |                                      |
| 5   | Get uri list by day and category           | gsi_date:date=`<date>` and begins_with(gsi_date:sk, `<category>`) |                                      |
|     |                                            |                                                                   |                                      |
