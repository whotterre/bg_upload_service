# Upload Service With Background Image Processing

A high-performance image upload service built with Go, featuring asynchronous background processing for image optimization and thumbnail generation.

## Overview

This service provides a REST API for uploading images with background processing capabilities. Images are stored in AWS S3, and processing tasks (resize, compress, thumbnail generation) are handled asynchronously using Asynq workers.

## Tech Stack

- **Go** - Primary programming language
- **Gin** - HTTP web framework
- **Asynq** - Distributed task queue (Redis-backed)
- **AWS S3** - Object storage for images
- **Redis** - Message broker and job queue
- **Image Processing** - imaging/disintegration libraries

## Features

- ✅ Async image upload to S3
- ✅ Background job processing with Asynq
- ✅ Image resizing and compression
- ✅ Automatic thumbnail generation
- ✅ Job status tracking
- ✅ Processed image URL retrieval

## Architecture

```                
┌─────────┐      ┌────────────┐      ┌───────┐
│ Client  │─────▶│  Gin API   │─────▶│  S3   │
└─────────┘      └────────────┘      └───────┘
                       │
                       ▼
                  ┌────────┐
                  │ Redis  │
                  └────────┘
                       │
                       ▼
              ┌────────────────┐
              │ Asynq Worker   │
              │ - Resize       │
              │ - Compress     │
              │ - Thumbnail    │
              └────────────────┘
                       │
                       ▼
                  ┌───────┐
                  │  S3   │
                  └───────┘
```

## Prerequisites

- Go 1.21 or higher
- Redis 6.0 or higher
- AWS Account with S3 access
- AWS credentials configured

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/bg_upload_service.git
cd bg_upload_service
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables:
```bash
cp .env.example .env
```

Edit `.env` with your configuration:
```env
# Server Configuration
PORT=8080
GIN_MODE=release

# Redis Configuration
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# AWS S3 Configuration
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
S3_BUCKET_NAME=your-bucket-name

# Image Processing
MAX_UPLOAD_SIZE=10485760  # 10MB
THUMBNAIL_WIDTH=200
THUMBNAIL_HEIGHT=200
RESIZE_WIDTH=1920
RESIZE_HEIGHT=1080
COMPRESS_QUALITY=85
```

## Running the Service

### Start the API Server

```bash
go run cmd/api/main.go
```

### Using Docker Compose

```bash
docker-compose up -d
```

## API Endpoints

### 1. Upload Image

Upload an image file for background processing.

**Endpoint:** `POST /upload`

**Request:**
- Content-Type: `multipart/form-data`
- Field: `image` (file)

**Example:**
```bash
curl -X POST http://localhost:8080/upload \
  -F "image=@/path/to/image.jpg"
```

**Response:**
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "pending",
  "message": "Image uploaded and queued for processing",
  "links": {
    "status": "/upload/550e8400-e29b-41d4-a716-446655440000/status",
    "result": "/upload/550e8400-e29b-41d4-a716-446655440000/result"
  }
}
```

### 2. Check Upload Status

Get the current status of an upload job.

**Endpoint:** `GET /upload/{id}/status`

**Example:**
```bash
curl http://localhost:8080/upload/550e8400-e29b-41d4-a716-446655440000/status
```

**Response (Processing):**
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "processing",
  "progress": 50,
  "error_message": "",
  "created_at": "2025-12-06T10:00:00Z",
  "completed_at": null
}
```

**Response (Completed):**
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "progress": 100,
  "error_message": "",
  "created_at": "2025-12-06T10:00:00Z",
  "completed_at": "2025-12-06T10:00:05Z"
}
```

### 3. Get Processed Results

Retrieve URLs for all processed image variants.

**Endpoint:** `GET /upload/{id}/result`

**Example:**
```bash
curl http://localhost:8080/upload/550e8400-e29b-41d4-a716-446655440000/result
```

**Response:**
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "completed",
  "original": {
    "url": "https://your-bucket.s3.amazonaws.com/originals/550e8400.jpg",
    "size": 2048000
  },
  "processed": {
    "resized_url": "https://your-bucket.s3.amazonaws.com/resized/550e8400.jpg",
    "compressed_url": "https://your-bucket.s3.amazonaws.com/compressed/550e8400.jpg",
    "thumbnail_url": "https://your-bucket.s3.amazonaws.com/thumbnails/550e8400.jpg",
    "compressed_size": 512000
  },
  "metadata": {
    "width": 1920,
    "height": 1080,
    "format": "jpeg"
  }
}
```

## Project Structure

```
bg_upload_service/
├── cmd/
│   ├── api/
│   │   └── main.go           # API server entry point
│   └── worker/
│       └── main.go           # Worker entry point
├── internal/
│   ├── config/
│   │   └── config.go         # Configuration management
│   ├── handlers/
│   │   └── upload.go         # HTTP handlers
│   ├── models/
│   │   └── upload.go         # Data models
│   ├── services/
│   │   ├── s3.go            # S3 upload service
│   │   └── processor.go     # Image processing service
│   ├── tasks/
│   │   └── image_tasks.go   # Asynq task definitions
│   └── workers/
│       └── image_worker.go  # Asynq worker implementation
├── pkg/
│   └── utils/
│       └── logger.go         # Logging utilities
├── .env.example
├── .gitignore
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```

## Image Processing Pipeline

1. **Upload Phase**
   - Validate image format (JPEG, PNG, GIF, WebP)
   - Check file size limits
   - Upload original to S3
   - Enqueue processing job

2. **Processing Phase** (Background Worker)
   - **Resize**: Scale image to configured dimensions while maintaining aspect ratio
   - **Compress**: Reduce file size with quality settings
   - **Thumbnail**: Generate small preview image
   - Upload processed variants to S3
   - Update job status

3. **Result Phase**
   - Return URLs for all processed variants
   - Include metadata (sizes, dimensions, format)

## Error Handling

The service handles various error scenarios:

- **Invalid file format**: Returns 400 with error message
- **File too large**: Returns 413 Payload Too Large
- **S3 upload failure**: Job marked as failed, retry attempted
- **Processing errors**: Logged and job marked as failed
- **Network issues**: Automatic retry with exponential backoff

## Monitoring

### Health Check

```bash
curl http://localhost:8080/health
```

### Asynq Dashboard

Access the Asynq web UI for monitoring:
```bash
asynq stats
asynq queue ls
```

## Development

### Run Tests

```bash
go test ./...
```

### Run with Live Reload

```bash
go install github.com/cosmtrek/air@latest
air
```

### Linting

```bash
golangci-lint run
```

## Performance Considerations

- **Concurrent Workers**: Configure multiple worker goroutines for parallel processing
- **Redis Connection Pool**: Optimized for high throughput
- **S3 Multipart Upload**: Used for large files
- **Image Streaming**: Process images without loading entirely into memory
- **Rate Limiting**: Implement to prevent API abuse

## Security

- ✅ File type validation (magic number check)
- ✅ File size limits
- ✅ AWS IAM role-based access
- ✅ Pre-signed URLs for temporary access
- ✅ Input sanitization
- ⚠️ TODO: Add authentication/authorization
- ⚠️ TODO: Add rate limiting

## Environment-Specific Configurations

### Development
```bash
GIN_MODE=debug
```

### Production
```bash
GIN_MODE=release
```

## Troubleshooting

### Worker not processing jobs
- Check Redis connection: `redis-cli ping`
- Verify worker is running
- Check worker logs for errors

### Images not uploading to S3
- Verify AWS credentials
- Check S3 bucket permissions
- Ensure bucket exists and is accessible

### Out of memory errors
- Reduce concurrent worker count
- Increase available memory
- Implement streaming for large files

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For issues and questions:
- Create an issue on GitHub
- Email: support@yourservice.com

## Roadmap

- [ ] Add support for video processing
- [ ] Implement webhook notifications
- [ ] Add batch upload support
- [ ] Support for MinIO (local S3-compatible storage)
- [ ] Add image watermarking
- [ ] Implement CDN integration
- [ ] Add WebSocket support for real-time status updates
- [ ] Support for additional image formats (HEIC, AVIF)

## Acknowledgments

- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [Asynq](https://github.com/hibiken/asynq)
- [AWS SDK for Go](https://github.com/aws/aws-sdk-go)
- [Imaging](https://github.com/disintegration/imaging)
