# TechYouth-Be API Documentation

## Overview

TechYouth-Be is a RESTful API for a historical learning game application. The API provides endpoints for user authentication, content management, progress tracking, and social features.

**Base URL:** `http://localhost:8000`  
**API Version:** v1  
**Content-Type:** `application/json`

## Authentication

Most endpoints require Bearer token authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your_token>
```

## Response Format

All API responses follow this standard format:

```json
{
  "code": 200,
  "message": "Success",
  "data": {
    // Response data here
  }
}
```

## Error Handling

Error responses include appropriate HTTP status codes and error messages:

```json
{
  "code": 400,
  "message": "Invalid request",
  "data": null
}
```

---

## Endpoints

### Health Check

#### GET /ping
Check the health status of the API service.

**Response:**
```json
{
  "code": 200,
  "message": "Success",
  "data": "pong"
}
```

---

## Authentication Endpoints

### POST /api/v1/register
Register a new user account.

**Request Body:**
```json
{
  "email": "user@example.com",
  "username": "username",
  "password": "password123"
}
```

**Response:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "user_id": "uuid-string"
  }
}
```

### POST /api/v1/login
Authenticate user and receive access token.

**Request Body:**
```json
{
  "email_or_username": "user@example.com",
  "password": "password123"
}
```

**Response:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "access_token": "jwt-token-string"
  }
}
```

### GET /api/v1/username/check/{username}
Check if a username is available for registration.

**Parameters:**
- `username` (path): Username to check

**Response:**
```json
{
  "code": 200,
  "message": "Username is available",
  "data": {
    "available": true,
    "username": "testuser"
  }
}
```

---

## Guest Session Endpoints

### POST /api/v1/guest/session
Create or retrieve a guest session for anonymous users.

**Request Body:**
```json
{
  "device_id": "unique-device-identifier"
}
```

**Response:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "session": {
      "id": "session-uuid",
      "device_id": "device-id",
      "created_at": "2023-01-01T00:00:00Z"
    },
    "progress": {
      "hearts": 5,
      "completed_lessons": [],
      "unlocked_characters": []
    }
  }
}
```

### GET /api/v1/guest/session/{sessionId}/progress
Get progress for a guest session.

**Parameters:**
- `sessionId` (path): Guest session ID

**Response:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "hearts": 5,
    "max_hearts": 5,
    "completed_lessons": ["lesson-1", "lesson-2"],
    "unlocked_characters": ["character-1"]
  }
}
```

### GET /api/v1/guest/session/{sessionId}/lesson/{lessonId}/access
Check if guest can access a specific lesson.

**Parameters:**
- `sessionId` (path): Guest session ID
- `lessonId` (path): Lesson ID to check

**Response:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "can_access": true,
    "reason": "Access granted",
    "hearts_needed": 0
  }
}
```

### POST /api/v1/guest/session/{sessionId}/lesson/complete
Mark a lesson as completed for guest session.

**Parameters:**
- `sessionId` (path): Guest session ID

**Request Body:**
```json
{
  "lesson_id": "lesson-uuid",
  "score": 85,
  "time_spent": 300
}
```

**Response:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "hearts": 5,
    "completed_lessons": ["lesson-1", "lesson-2"],
    "unlocked_characters": ["character-1"]
  }
}
```

### POST /api/v1/guest/session/{sessionId}/hearts/add
Add hearts to guest session (from watching ads).

**Parameters:**
- `sessionId` (path): Guest session ID

**Response:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "hearts": 5,
    "max_hearts": 5
  }
}
```

### POST /api/v1/guest/session/{sessionId}/hearts/lose
Deduct a heart from guest session (failed lesson).

**Parameters:**
- `sessionId` (path): Guest session ID

**Response:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "hearts": 4,
    "max_hearts": 5
  }
}
```

---

## Content Endpoints

### GET /api/v1/content/timeline
Get the historical timeline with eras and dynasties.

**Response:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "eras": [
      {
        "era": "Ancient Period",
        "dynasties": [
          {
            "dynasty": "Hồng Bàng",
            "start_year": -2879,
            "end_year": -258,
            "characters": [],
            "is_unlocked": true,
            "progress": 0.5
          }
        ],
        "is_unlocked": true,
        "progress": 0.3
      }
    ]
  }
}
```

### GET /api/v1/content/characters
Get list of historical characters with optional filtering.

**Query Parameters:**
- `dynasty` (optional): Filter by dynasty name
- `rarity` (optional): Filter by rarity level

**Response:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "characters": [
      {
        "id": "character-uuid",
        "name": "Lý Thái Tổ",
        "dynasty": "Lý Dynasty",
        "rarity": "legendary",
        "birth_year": 974,
        "death_year": 1028,
        "description": "Founder of the Lý Dynasty",
        "famous_quote": "Famous historical quote",
        "achievements": ["Founded dynasty", "Established capital"],
        "image_url": "https://example.com/image.jpg",
        "is_unlocked": true,
        "lesson_count": 5
      }
    ],
    "total": 50,
    "unlocked": 10
  }
}
```

### GET /api/v1/content/characters/{characterId}
Get detailed information about a specific character.

**Parameters:**
- `characterId` (path): Character ID

**Response:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "id": "character-uuid",
    "name": "Lý Thái Tổ",
    "dynasty": "Lý Dynasty",
    "rarity": "legendary",
    "birth_year": 974,
    "death_year": 1028,
    "description": "Detailed character biography...",
    "famous_quote": "Historical quote",
    "achievements": ["Achievement 1", "Achievement 2"],
    "image_url": "https://example.com/image.jpg",
    "is_unlocked": true,
    "lesson_count": 5
  }
}
```

### GET /api/v1/content/characters/{characterId}/lessons
Get all lessons for a specific character.

**Parameters:**
- `characterId` (path): Character ID

**Response:**
```json
{
  "code": 200,
  "message": "Success",
  "data": [
    {
      "id": "lesson-uuid",
      "character_id": "character-uuid",
      "title": "The Rise of Lý Thái Tổ",
      "order": 1,
      "story": "Lesson story content...",
      "voice_over_url": "https://example.com/audio.mp3",
      "questions": [
        {
          "id": "question-uuid",
          "type": "multiple_choice",
          "question": "When was Lý Thái Tổ born?",
          "options": ["974", "975", "976", "977"],
          "points": 10
        }
      ],
      "xp_reward": 100,
      "min_score": 70,
      "character": {
        "id": "character-uuid",
        "name": "Lý Thái Tổ"
      }
    }
  ]
}
```

### GET /api/v1/content/lessons/{lessonId}
Get detailed lesson content including questions.

**Parameters:**
- `lessonId` (path): Lesson ID

**Response:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "id": "lesson-uuid",
    "character_id": "character-uuid",
    "title": "Lesson Title",
    "order": 1,
    "story": "Lesson story content...",
    "voice_over_url": "https://example.com/audio.mp3",
    "questions": [
      {
        "id": "question-uuid",
        "type": "multiple_choice",
        "question": "Question text?",
        "options": ["Option 1", "Option 2", "Option 3", "Option 4"],
        "points": 10,
        "metadata": {}
      }
    ],
    "xp_reward": 100,
    "min_score": 70,
    "character": {}
  }
}
```

### GET /api/v1/content/search
Search characters and content with filters.

**Query Parameters:**
- `query` (optional): Search query string
- `dynasty` (optional): Filter by dynasty
- `rarity` (optional): Filter by rarity
- `limit` (optional): Limit results (default: 20)

**Response:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "characters": [
      {
        "id": "character-uuid",
        "name": "Character Name",
        "dynasty": "Dynasty Name",
        "rarity": "common"
      }
    ],
    "total": 5
  }
}
```

---

## User Endpoints (Authenticated)

### GET /api/v1/user/profile
Get current user's profile information.

**Headers:**
```
Authorization: Bearer <token>
```

**Response:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "user_id": "user-uuid",
    "email": "user@example.com",
    "username": "username",
    "birth_year": 1990,
    "joined_at": "2023-01-01T00:00:00Z",
    "last_login_at": "2023-01-15T10:30:00Z"
  }
}
```

### PUT /api/v1/user/profile
Update user profile information.

**Headers:**
```
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "username": "new_username",
  "birth_year": 1990
}
```

**Response:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "user_id": "user-uuid",
    "email": "user@example.com",
    "username": "new_username",
    "birth_year": 1990,
    "joined_at": "2023-01-01T00:00:00Z",
    "last_login_at": "2023-01-15T10:30:00Z"
  }
}
```

### POST /api/v1/user/initialize
Initialize user profile after first login (zodiac setup).

**Headers:**
```
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "birth_year": 1990
}
```

**Response:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "user_id": "user-uuid",
    "hearts": 5,
    "max_hearts": 5,
    "xp": 0,
    "level": 1,
    "xp_to_next_level": 100,
    "completed_lessons": [],
    "unlocked_characters": [],
    "streak": 0,
    "total_play_time": 0,
    "spirit": {
      "id": "spirit-uuid",
      "type": "dragon",
      "stage": 1,
      "xp": 0,
      "xp_to_next": 100,
      "name": "Baby Dragon",
      "image_url": "https://example.com/spirit.jpg"
    },
    "recent_achievements": []
  }
}
```

### GET /api/v1/user/progress
Get current user's game progress.

**Headers:**
```
Authorization: Bearer <token>
```

**Response:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "user_id": "user-uuid",
    "hearts": 5,
    "max_hearts": 5,
    "xp": 1500,
    "level": 5,
    "xp_to_next_level": 200,
    "completed_lessons": ["lesson-1", "lesson-2"],
    "unlocked_characters": ["character-1", "character-2"],
    "streak": 7,
    "total_play_time": 3600,
    "last_heart_reset": "2023-01-15T00:00:00Z",
    "last_activity": "2023-01-15T10:30:00Z",
    "spirit": {
      "id": "spirit-uuid",
      "type": "dragon",
      "stage": 2,
      "xp": 150,
      "xp_to_next": 50,
      "name": "Young Dragon",
      "image_url": "https://example.com/spirit.jpg"
    },
    "recent_achievements": []
  }
}
```

### GET /api/v1/user/stats
Get detailed user statistics and analytics.

**Headers:**
```
Authorization: Bearer <token>
```

**Response:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "user_id": "user-uuid",
    "level": 5,
    "xp": 1500,
    "hearts": 5,
    "streak": 7,
    "total_play_time": 3600,
    "completed_lessons": 25,
    "unlocked_characters": 10,
    "achievements": 5,
    "rank": 150,
    "spirit_stage": 2,
    "spirit_type": "dragon",
    "weekly_stats": {
      "lessons_completed": 5,
      "xp_gained": 500
    },
    "monthly_stats": {
      "lessons_completed": 20,
      "xp_gained": 2000
    }
  }
}
```

### GET /api/v1/user/collection
Get user's character collection and achievements.

**Headers:**
```
Authorization: Bearer <token>
```

**Query Parameters:**
- `userId` (optional): View another user's collection

**Response:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "characters": {
      "characters": [],
      "total": 50,
      "unlocked": 10
    },
    "achievements": [
      {
        "id": "achievement-uuid",
        "name": "First Steps",
        "description": "Complete your first lesson",
        "badge_url": "https://example.com/badge.jpg",
        "category": "progress",
        "xp_reward": 50,
        "unlocked_at": "2023-01-01T00:00:00Z"
      }
    ],
    "stats": {
      "total_characters": 50,
      "unlocked_characters": 10,
      "completion_rate": 0.2,
      "rarity_breakdown": {
        "common": 5,
        "rare": 3,
        "legendary": 2
      },
      "dynasty_breakdown": {
        "Lý Dynasty": 4,
        "Trần Dynasty": 3,
        "Lê Dynasty": 3
      }
    }
  }
}
```

### GET /api/v1/user/lesson/{lessonId}/access
Check if user can access a specific lesson.

**Headers:**
```
Authorization: Bearer <token>
```

**Parameters:**
- `lessonId` (path): Lesson ID

**Response:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "can_access": true,
    "reason": "Access granted",
    "hearts_needed": 0
  }
}
```

### POST /api/v1/user/lesson/complete
Complete a lesson for registered user.

**Headers:**
```
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "lesson_id": "lesson-uuid",
  "score": 85,
  "time_spent": 300
}
```

**Response:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "xp_gained": 100,
    "new_level": 5,
    "leveled_up": true,
    "character_unlock": "character-uuid",
    "spirit_evolved": false
  }
}
```

### GET /api/v1/user/hearts
Get current heart status and reset information.

**Headers:**
```
Authorization: Bearer <token>
```

**Response:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "hearts": 5,
    "max_hearts": 5,
    "next_reset_time": "2023-01-16T00:00:00Z",
    "can_watch_ad": true,
    "ads_watched_today": 2
  }
}
```

### POST /api/v1/user/hearts/add
Add hearts from ads or other sources.

**Headers:**
```
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "source": "ad",
  "amount": 1
}
```

**Response:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "hearts": 5,
    "max_hearts": 5,
    "next_reset_time": "2023-01-16T00:00:00Z",
    "can_watch_ad": false,
    "ads_watched_today": 3
  }
}
```

### POST /api/v1/user/hearts/lose
Deduct a heart when user fails a lesson.

**Headers:**
```
Authorization: Bearer <token>
```

**Response:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "hearts": 4,
    "max_hearts": 5,
    "next_reset_time": "2023-01-16T00:00:00Z",
    "can_watch_ad": true,
    "ads_watched_today": 2
  }
}
```

### POST /api/v1/user/share
Share user achievement or progress on social media.

**Headers:**
```
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "type": "achievement",
  "content": "Just unlocked a new character!",
  "item_id": "character-uuid"
}
```

**Response:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "share_url": "https://techyouth.com/share/abc123",
    "share_image": "https://example.com/share-image.jpg",
    "share_text": "I just unlocked Lý Thái Tổ in TechYouth!",
    "platforms": ["facebook", "twitter", "instagram"]
  }
}
```

---

## Leaderboard Endpoints

### GET /api/v1/leaderboard/weekly
Get weekly leaderboard rankings.

**Query Parameters:**
- `limit` (optional): Limit results (default: 50, max: 100)

**Response:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "period": "weekly",
    "current_user": {
      "user_id": "user-uuid",
      "username": "current_user",
      "level": 5,
      "xp": 1500,
      "rank": 25,
      "spirit_type": "dragon",
      "spirit_stage": 2
    },
    "top_users": [
      {
        "user_id": "top-user-uuid",
        "username": "top_player",
        "level": 10,
        "xp": 5000,
        "rank": 1,
        "spirit_type": "phoenix",
        "spirit_stage": 3
      }
    ]
  }
}
```

### GET /api/v1/leaderboard/monthly
Get monthly leaderboard rankings.

**Query Parameters:**
- `limit` (optional): Limit results (default: 50, max: 100)

**Response:** Same format as weekly leaderboard

### GET /api/v1/leaderboard/all-time
Get all-time leaderboard rankings.

**Query Parameters:**
- `limit` (optional): Limit results (default: 50, max: 100)

**Response:** Same format as weekly leaderboard

---

## Admin Endpoints (Future Implementation)

### POST /api/v1/admin/characters
Create a new historical character (admin only).

**Headers:**
```
Authorization: Bearer <admin_token>
```

**Request Body:**
```json
{
  "name": "Character Name",
  "dynasty": "Dynasty Name",
  "rarity": "legendary",
  "birth_year": 1000,
  "death_year": 1050,
  "description": "Character description",
  "famous_quote": "Famous quote",
  "achievements": ["Achievement 1", "Achievement 2"],
  "image_url": "https://example.com/image.jpg"
}
```

### POST /api/v1/admin/lessons
Create a new lesson (admin only).

**Headers:**
```
Authorization: Bearer <admin_token>
```

**Request Body:**
```json
{
  "character_id": "character-uuid",
  "title": "Lesson Title",
  "order": 1,
  "story": "Lesson story content",
  "voice_over_url": "https://example.com/audio.mp3",
  "questions": [],
  "xp_reward": 100,
  "min_score": 70
}
```

---

## Status Codes

- `200 OK` - Request successful
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid request parameters
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Access denied
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

---

## Rate Limiting

The API implements rate limiting to prevent abuse. Limits are applied per IP address and authenticated user.

---

## Swagger Documentation

Interactive API documentation is available at:
```
GET /swagger/index.html
```

This provides a complete interactive interface for testing all endpoints.
