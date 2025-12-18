// Package docs contains the Swagger documentation for the Arabella API
package docs

import "github.com/swaggo/swag"

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0.0",
	Host:             "api.arabella.uz",
	BasePath:         "/api/v1",
	Schemes:          []string{"http", "https"},
	Title:            "Arabella API",
	Description:      "AI Video Generation Platform API - Create professional AI-generated videos using pre-built templates",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

const docTemplate = `{
    "swagger": "2.0",
    "info": {
        "title": "Arabella API",
        "description": "AI Video Generation Platform API - Create professional AI-generated videos using pre-built templates",
        "termsOfService": "https://arabella.app/terms",
        "contact": {
            "name": "API Support",
            "url": "https://arabella.app/support",
            "email": "support@arabella.app"
        },
        "license": {
            "name": "MIT",
            "url": "https://opensource.org/licenses/MIT"
        },
        "version": "1.0.0"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/auth/google": {
            "post": {
                "description": "Authenticate a user using Google OAuth ID token",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["auth"],
                "summary": "Authenticate with Google",
                "parameters": [
                    {
                        "description": "Google ID Token",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/GoogleAuthRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/GoogleAuthResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            }
        },
        "/auth/refresh": {
            "post": {
                "description": "Refresh an access token using a refresh token",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["auth"],
                "summary": "Refresh access token",
                "parameters": [
                    {
                        "description": "Refresh Token",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/RefreshTokenRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/AuthTokens"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            }
        },
        "/templates": {
            "get": {
                "description": "Get a paginated list of video templates",
                "produces": ["application/json"],
                "tags": ["templates"],
                "summary": "List templates",
                "parameters": [
                    {"name": "category", "in": "query", "type": "string"},
                    {"name": "search", "in": "query", "type": "string"},
                    {"name": "premium", "in": "query", "type": "boolean"},
                    {"name": "page", "in": "query", "type": "integer", "default": 1},
                    {"name": "page_size", "in": "query", "type": "integer", "default": 20}
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/TemplateListResponse"
                        }
                    }
                }
            }
        },
        "/templates/{id}": {
            "get": {
                "description": "Get a video template by ID",
                "produces": ["application/json"],
                "tags": ["templates"],
                "summary": "Get template",
                "parameters": [
                    {"name": "id", "in": "path", "required": true, "type": "string", "format": "uuid"}
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/Template"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            }
        },
        "/videos/generate": {
            "post": {
                "security": [{"BearerAuth": []}],
                "description": "Start a new AI video generation job",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["videos"],
                "summary": "Generate video",
                "parameters": [
                    {
                        "description": "Video Generation Request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/GenerateVideoRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/VideoGenerationResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    },
                    "403": {
                        "description": "Forbidden",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            }
        },
        "/videos/{id}/status": {
            "get": {
                "security": [{"BearerAuth": []}],
                "description": "Get the current status of a video generation job",
                "produces": ["application/json"],
                "tags": ["videos"],
                "summary": "Get job status",
                "parameters": [
                    {"name": "id", "in": "path", "required": true, "type": "string", "format": "uuid"}
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/VideoJob"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            }
        },
        "/user/profile": {
            "get": {
                "security": [{"BearerAuth": []}],
                "description": "Get the profile of the authenticated user",
                "produces": ["application/json"],
                "tags": ["user"],
                "summary": "Get user profile",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/UserProfileResponse"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "ErrorResponse": {
            "type": "object",
            "properties": {
                "error": {"type": "string"},
                "code": {"type": "string"},
                "details": {"type": "string"}
            }
        },
        "GoogleAuthRequest": {
            "type": "object",
            "required": ["id_token"],
            "properties": {
                "id_token": {"type": "string"}
            }
        },
        "GoogleAuthResponse": {
            "type": "object",
            "properties": {
                "user": {"$ref": "#/definitions/User"},
                "tokens": {"$ref": "#/definitions/AuthTokens"}
            }
        },
        "RefreshTokenRequest": {
            "type": "object",
            "required": ["refresh_token"],
            "properties": {
                "refresh_token": {"type": "string"}
            }
        },
        "AuthTokens": {
            "type": "object",
            "properties": {
                "access_token": {"type": "string"},
                "refresh_token": {"type": "string"},
                "expires_at": {"type": "string", "format": "date-time"},
                "token_type": {"type": "string"}
            }
        },
        "User": {
            "type": "object",
            "properties": {
                "id": {"type": "string", "format": "uuid"},
                "email": {"type": "string"},
                "name": {"type": "string"},
                "avatar_url": {"type": "string"},
                "credits": {"type": "integer"},
                "tier": {"type": "string", "enum": ["free", "premium", "pro"]},
                "subscription_expires_at": {"type": "string", "format": "date-time"},
                "created_at": {"type": "string", "format": "date-time"}
            }
        },
        "Template": {
            "type": "object",
            "properties": {
                "id": {"type": "string", "format": "uuid"},
                "name": {"type": "string"},
                "category": {"type": "string"},
                "description": {"type": "string"},
                "thumbnail_url": {"type": "string"},
                "credit_cost": {"type": "integer"},
                "estimated_time": {"type": "integer"},
                "is_premium": {"type": "boolean"}
            }
        },
        "TemplateListResponse": {
            "type": "object",
            "properties": {
                "templates": {"type": "array", "items": {"$ref": "#/definitions/Template"}},
                "total": {"type": "integer"},
                "page": {"type": "integer"},
                "page_size": {"type": "integer"},
                "total_pages": {"type": "integer"}
            }
        },
        "GenerateVideoRequest": {
            "type": "object",
            "required": ["template_id", "prompt"],
            "properties": {
                "template_id": {"type": "string", "format": "uuid"},
                "prompt": {"type": "string", "minLength": 10, "maxLength": 2000},
                "params": {"$ref": "#/definitions/VideoParams"}
            }
        },
        "VideoParams": {
            "type": "object",
            "properties": {
                "duration": {"type": "integer"},
                "resolution": {"type": "string", "enum": ["720p", "1080p", "4k"]},
                "aspect_ratio": {"type": "string", "enum": ["16:9", "9:16", "1:1"]},
                "fps": {"type": "integer"}
            }
        },
        "VideoGenerationResponse": {
            "type": "object",
            "description": "Response after starting a video generation job",
            "properties": {
                "job_id": {"type": "string", "format": "uuid", "example": "550e8400-e29b-41d4-a716-446655440000"},
                "status": {"type": "string", "enum": ["pending", "processing", "diffusing", "uploading", "completed", "failed", "cancelled"], "example": "pending"},
                "estimated_time": {"type": "integer", "description": "Estimated time in seconds", "example": 90},
                "queue_position": {"type": "integer", "description": "Position in the generation queue", "example": 0}
            }
        },
        "VideoJob": {
            "type": "object",
            "description": "Video generation job with status, progress, and result URLs",
            "properties": {
                "id": {"type": "string", "format": "uuid", "example": "550e8400-e29b-41d4-a716-446655440000"},
                "user_id": {"type": "string", "format": "uuid", "example": "550e8400-e29b-41d4-a716-446655440000"},
                "template_id": {"type": "string", "format": "uuid", "example": "550e8400-e29b-41d4-a716-446655440000"},
                "prompt": {"type": "string", "example": "A beautiful sunset over mountains"},
                "params": {"$ref": "#/definitions/VideoParams"},
                "status": {"type": "string", "enum": ["pending", "processing", "diffusing", "uploading", "completed", "failed", "cancelled"], "example": "completed"},
                "progress": {"type": "integer", "minimum": 0, "maximum": 100, "example": 100},
                "provider": {"type": "string", "enum": ["gemini_veo", "openai_sora", "runway", "pika_labs", "mock"], "example": "gemini_veo"},
                "provider_job_id": {"type": "string", "example": "gemini-job-123"},
                "video_url": {"type": "string", "example": "https://storage.googleapis.com/gemini-videos/abc123.mp4"},
                "thumbnail_url": {"type": "string", "example": "https://cdn.arabella.app/thumbnails/abc123.jpg"},
                "duration_seconds": {"type": "integer", "example": 15},
                "credits_charged": {"type": "integer", "example": 2},
                "error_message": {"type": "string", "example": "Generation failed"},
                "created_at": {"type": "string", "format": "date-time", "example": "2025-12-13T16:00:00Z"},
                "started_at": {"type": "string", "format": "date-time", "example": "2025-12-13T16:00:05Z"},
                "completed_at": {"type": "string", "format": "date-time", "example": "2025-12-13T16:02:00Z"}
            }
        },
        "UserProfileResponse": {
            "type": "object",
            "properties": {
                "user": {"$ref": "#/definitions/User"},
                "active_jobs_count": {"type": "integer"},
                "total_jobs": {"type": "integer"}
            }
        }
    },
    "securityDefinitions": {
        "BearerAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header",
            "description": "Enter the token with the ` + "`" + `Bearer ` + "`" + ` prefix"
        }
    }
}`

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}

