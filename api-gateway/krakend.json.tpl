{
  "version": 3,
  "name": "API Gateway",
  "port": 8080,
  "timeout": "3s",
  "cache_ttl": "300s",
  "output_encoding": "json",
  "endpoints": [
    {
      "endpoint": "/orders",
      "method": "POST",
      "backend": [
        {
          "host": [
            "$ORDER_SERVICE_URL"
          ],
          "url_pattern": "/api/orders",
          "extra_config": {
            "backend/http": {
              "return_error_code": true
            }
          }
        }
      ]
    },
    {
      "endpoint": "/orders/{orderID}",
      "method": "GET",
      "backend": [
        {
          "host": [
            "$ORDER_SERVICE_URL"
          ],
          "url_pattern": "/api/orders/{orderID}",
          "extra_config": {
            "backend/http": {
              "return_error_code": true
            }
          }
        }
      ]
    },
    {
      "endpoint": "/users",
      "method": "POST",
      "backend": [
        {
          "host": [
            "$USER_SERVICE_URL"
          ],
          "url_pattern": "/api/users",
          "extra_config": {
            "backend/http": {
              "return_error_code": true
            }
          }
        }
      ]
    },
    {
      "endpoint": "/users",
      "method": "GET",
      "backend": [
        {
          "host": [
            "$USER_SERVICE_URL"
          ],
          "url_pattern": "/api/users",
          "extra_config": {
            "backend/http": {
              "return_error_code": true
            }
          }
        }
      ]
    },
    {
      "endpoint": "/products",
      "method": "POST",
      "backend": [
        {
          "host": [
            "$INVENTORY_SERVICE_URL"
          ],
          "url_pattern": "/api/products",
          "extra_config": {
            "backend/http": {
              "return_error_code": true
            }
          }
        }
      ]
    },
    {
      "endpoint": "/products/{id}",
      "method": "GET",
      "backend": [
        {
          "host": [
            "$INVENTORY_SERVICE_URL"
          ],
          "url_pattern": "/api/products/{id}",
          "extra_config": {
            "backend/http": {
              "return_error_code": true
            }
          }
        }
      ]
    },
    {
      "endpoint": "/products",
      "method": "GET",
      "backend": [
        {
          "host": [
            "$INVENTORY_SERVICE_URL"
          ],
          "url_pattern": "/api/products",
          "extra_config": {
            "backend/http": {
              "return_error_code": true
            }
          }
        }
      ]
    }
  ]
}
