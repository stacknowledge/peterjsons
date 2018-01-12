Peterjsons helps dealing with random json input when the usecase is a transformation to another json output. It receives a random json input and a json recipe to "cook" the json output.

## How to install:

```bash
go get github.com/stacknowledge/peterjsons
```

## How to use:

```golang
package main

import (
	"fmt"

	"github.com/stacknowledge/peterjsons"
)

func main() {
	material := []byte(`
            {
                "info" : {
                    "peter":"jsons"
                }
            }
        `)

	recipe := []byte(`
            {
                "recipe": {
                    "id" : "info.peter"
                }
            }
        `)

    pjsons, err := peterjsons.New(material, recipe)
	if err != nil {
		fmt.Printf("\n %s \n", err)
    }
    
	pjsons.Cook()

    fmt.Printf("\n %s \n", pjsons.JSONResult())

    // ------------------------------------------------
    //             Outputs : {"id":"jsons"}
    // ------------------------------------------------
}
```

## Example:

### Input example
```json
{
  "data": {
    "product": {
      "id": "identification",
      "title": "title",
      "description": "This is description and now it has more than 20 chars.",
      "version": 1,
      "category_code": "category",
      "contact": {
        "name": "peters",
        "phones": ["790123123", "790123546"],
        "logo": "https://peters.logos.com/jsons.jpg"
      },
      "properties": {
        "multiple": [
          {
            "code": "peter",
            "value": "json"
          },
          {
            "code": "json",
            "value": "peter"
          }
        ]
      }
    }
  }
}
```

### Recipe Example:

```json
 {
  "recipe": {
    "contact": "data.product.contact",
    "meta.id": "data.product.id",
    "products": {
      "values": ["data.product.properties.multiple"],
      "operation": "swapmap",
      "*": {
        "id": {
          "value": "code"
        },
        "description": {
          "value": "value"
        },
        "price": {
          "replace": 1
        }
      }
    },
    "info": {
      "resume": {
        "values": ["data.product.title", "data.product.description"],
        "operation": "concat",
        "separator": " - "
      },
      "category": {
        "values": ["data.product.category_code"],
        "operation": "swap",
        "replace": ["one_bed"]
      },
      "version": {
        "values": ["data.product.version"],
        "operation": "swap",
        "format": "v%v.0.0"
      },
      "advertiser": {
        "values": ["data.product.advertiser_type"],
        "operation": "swap",
        "replace": [1]
      }
    }
  }
}
```

### Result:
```json
{
  "meta": {
    "id": "identification"
  },
  "contact": {
    "logo": "https://peters.logos.com/jsons.jpg",
    "name": "peters",
    "phones": ["790123123", "790123546"]
  },
  "info": {
    "advertiser": 1,
    "category": "one_bed",
    "resume": "title - This is description and now it has more than 20 chars.",
    "version": "v1.0.0"
  },
  "products": [
    {
      "description": "json",
      "id": "peter",
      "price": 1
    },
    {
      "description": "peter",
      "id": "json",
      "price": 1
    }
  ]
}
```