# bolha-utils

## upload example
```sh
bolha-utils upload --file=ads.json
```

## ads.json example
```json
[
    {
        "user": {
            "username": "johndoe",
            "password": "password123"
        },
        "ads": [
            {
                "title": "Lorem ipsum dolor sit amet",
                "description": "Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.",
                "price": "123",
                "categoryId": "1234",
                "images": [
                    "img1.jpg",
                    "img2.jpg"
                ]
            }
        ]
    }
]
```
