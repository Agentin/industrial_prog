# API Documentation for Practical Work 17

## Auth Service (port 8081)

### `POST /v1/auth/login`
Получение токена доступа (учебный).

**Request Body:**
```json
{
  "username": "student",
  "password": "student"
}