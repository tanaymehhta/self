# Self AI Services

Python-based AI processing services for audio transcription, NLP, and intelligent insights.

## Features

- **Local Transcription**: Whisper.cpp integration for fast, local speech-to-text
- **Entity Extraction**: spaCy NLP pipeline for named entity recognition
- **Semantic Search**: sentence-transformers for vector embeddings
- **LLM Integration**: Ollama for local language model inference
- **Pattern Recognition**: Custom algorithms for behavioral insights
- **Real-time Processing**: Async processing with NATS messaging

## Tech Stack

- **Language**: Python 3.11+
- **Framework**: FastAPI for HTTP API
- **Speech-to-Text**: Whisper.cpp (Python bindings)
- **NLP**: spaCy, Hugging Face Transformers
- **LLM**: Ollama (Llama 3.1, Mistral, CodeLlama)
- **Vector DB**: Qdrant for similarity search
- **Message Queue**: NATS for job processing
- **ML Libraries**: scikit-learn, numpy, pandas

## Getting Started

```bash
# Create virtual environment
python -m venv venv
source venv/bin/activate  # or `venv\Scripts\activate` on Windows

# Install dependencies
pip install -r requirements.txt

# Download language models
python scripts/download_models.py

# Start development server
python main.py

# Run with auto-reload
uvicorn main:app --reload --port 8000
```

## Environment Variables

Create `.env`:

```bash
# Message Queue
NATS_URL=localhost:4222

# Vector Database
QDRANT_URL=http://localhost:6333

# Model Paths
WHISPER_MODEL_PATH=./models/whisper-large-v3
SPACY_MODEL=en_core_web_lg
SENTENCE_TRANSFORMER_MODEL=all-MiniLM-L6-v2

# Ollama
OLLAMA_URL=http://localhost:11434
OLLAMA_MODEL=llama3.1:8b

# Hugging Face
HUGGINGFACE_HUB_CACHE=./models/huggingface
TRANSFORMERS_CACHE=./models/transformers

# Logging
LOG_LEVEL=INFO
LOG_FORMAT=json

# Performance
MAX_WORKERS=4
BATCH_SIZE=32
```

## Project Structure

```
ai-services/
├── app/                   # Application code
│   ├── api/              # FastAPI routes
│   ├── core/             # Core configuration
│   ├── models/           # Pydantic models
│   ├── services/         # Business logic
│   └── workers/          # Background workers
├── models/               # ML models (not in git)
├── scripts/              # Utility scripts
├── tests/                # Test files
└── requirements.txt      # Python dependencies
```

## API Endpoints

### Transcription
- `POST /api/v1/transcribe` - Transcribe audio file
- `POST /api/v1/transcribe/stream` - Real-time transcription
- `GET /api/v1/transcribe/:job_id` - Get transcription status

### NLP Processing
- `POST /api/v1/nlp/extract-entities` - Extract named entities
- `POST /api/v1/nlp/summarize` - Summarize text
- `POST /api/v1/nlp/classify` - Classify text content

### Semantic Search
- `POST /api/v1/embeddings/generate` - Generate text embeddings
- `POST /api/v1/search/semantic` - Semantic similarity search
- `POST /api/v1/search/hybrid` - Hybrid search (keyword + semantic)

### Insights
- `POST /api/v1/insights/analyze` - Analyze conversation patterns
- `GET /api/v1/insights/patterns` - Get behavioral patterns
- `POST /api/v1/insights/recommendations` - Generate recommendations

## Services

### TranscriptionService

```python
from app.services.transcription import TranscriptionService

service = TranscriptionService()

# Transcribe audio file
result = await service.transcribe_audio(
    audio_data=audio_bytes,
    language="en",
    model="whisper-large-v3"
)

# Real-time transcription
async for segment in service.transcribe_stream(audio_stream):
    print(f"[{segment.start:.2f}s] {segment.text}")
```

### NLPService

```python
from app.services.nlp import NLPService

service = NLPService()

# Extract entities
entities = await service.extract_entities(text)

# Generate summary
summary = await service.summarize(text, max_length=100)

# Classify content
classification = await service.classify(text, categories=["work", "personal"])
```

### EmbeddingService

```python
from app.services.embeddings import EmbeddingService

service = EmbeddingService()

# Generate embeddings
embeddings = await service.generate_embeddings([
    "First conversation about project planning",
    "Meeting discussion about budget allocation"
])

# Semantic search
results = await service.search_similar(
    query="project budget",
    embeddings=stored_embeddings,
    top_k=5
)
```

### InsightService

```python
from app.services.insights import InsightService

service = InsightService()

# Analyze patterns
patterns = await service.analyze_patterns(conversations)

# Generate insights
insights = await service.generate_insights(
    user_id="user123",
    time_window="7d"
)
```

## Background Workers

### Audio Processing Worker

```python
# Processes audio transcription jobs
python -m app.workers.audio_worker
```

### NLP Processing Worker

```python
# Handles entity extraction and text analysis
python -m app.workers.nlp_worker
```

### Insight Generation Worker

```python
# Generates proactive insights and recommendations
python -m app.workers.insight_worker
```

## Model Management

### Download Models

```bash
# Download all required models
python scripts/download_models.py

# Download specific model
python scripts/download_models.py --model whisper-large-v3

# List available models
python scripts/list_models.py
```

### Model Configuration

```python
# models/config.py
MODELS = {
    "whisper": {
        "small": "whisper-small",
        "medium": "whisper-medium",
        "large": "whisper-large-v3"
    },
    "spacy": {
        "english": "en_core_web_lg",
        "multilingual": "xx_ent_wiki_sm"
    },
    "embeddings": {
        "fast": "all-MiniLM-L6-v2",
        "accurate": "all-mpnet-base-v2"
    }
}
```

## Development

### Running Tests

```bash
# Run all tests
pytest

# Run with coverage
pytest --cov=app

# Run specific test file
pytest tests/test_transcription.py

# Run tests with specific marker
pytest -m integration
```

### Code Quality

```bash
# Format code
black .
isort .

# Lint code
flake8 .
mypy app/

# Check security
bandit -r app/
```

### Performance Profiling

```bash
# Profile API endpoints
python -m cProfile -o profile.stats main.py

# Memory profiling
python -m memory_profiler scripts/profile_transcription.py

# Line profiler
kernprof -l -v app/services/transcription.py
```

## Docker Deployment

### Development

```dockerfile
FROM python:3.11-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    ffmpeg \
    && rm -rf /var/lib/apt/lists/*

# Install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY . .

# Download models
RUN python scripts/download_models.py

EXPOSE 8000

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
```

### Production

```bash
# Build optimized image
docker build -t self-ai:latest .

# Run with GPU support
docker run --gpus all -p 8000:8000 self-ai:latest
```

## Performance Optimization

### GPU Acceleration

```python
# Enable CUDA for Whisper.cpp
os.environ["WHISPER_CPP_CUDA"] = "1"

# Use GPU for transformers
import torch
device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
```

### Batch Processing

```python
# Process multiple audio files in batches
batch_results = await service.transcribe_batch(
    audio_files=audio_batch,
    batch_size=8
)
```

### Caching

```python
# Cache embeddings and model outputs
from functools import lru_cache

@lru_cache(maxsize=1000)
def get_embedding(text: str) -> List[float]:
    return model.encode(text)
```

## Monitoring

### Health Checks

- `/health` - Service health status
- `/health/models` - Model loading status
- `/health/workers` - Worker status

### Metrics

- Processing time per request
- Model memory usage
- Queue length and processing rate
- Error rates by endpoint

### Logging

```python
import structlog

logger = structlog.get_logger()

logger.info(
    "transcription_completed",
    duration=processing_time,
    audio_length=audio_duration,
    model="whisper-large-v3"
)
```