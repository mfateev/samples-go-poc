### Steps to run this sample:

1. Follow [root setup](../README.md#setup), then start a server with the custom attributes below.

  ```
  temporal server start-dev \
    --search-attribute CustomIntField=Int \
    --search-attribute CustomKeywordField=Keyword \
    --search-attribute CustomBoolField=Bool \
    --search-attribute CustomDoubleField=Double \
    --search-attribute CustomDatetimeField=Datetime \
    --search-attribute CustomKeywordListField=KeywordList
  ```

1. Run the following command to start the worker:

  ```
  mkdir -p bin
  tgo build -o bin/searchattributes-worker searchattributes/worker/main.go
  ./bin/searchattributes-worker
  ```

1. Run the following command to start the example:

  ```
  tgo run searchattributes/starter/main.go
  ```

The worker logs workflow/run identity and the typed attribute values, including
large integers, keyword lists and history-clock timestamps. Native typed keys
use canonical name/value-type ordering rather than process descriptor addresses.
