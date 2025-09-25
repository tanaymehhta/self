# Contributing to Self

We welcome contributions to Self! This document provides guidelines for contributing to the project.

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct. Be respectful, inclusive, and professional in all interactions.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/yourusername/self.git`
3. Follow the [Development Guide](./DEVELOPMENT.md) to set up your environment
4. Create a feature branch: `git checkout -b feature/your-feature-name`

## Development Process

### Branch Naming

- `feature/description` - New features
- `fix/description` - Bug fixes
- `docs/description` - Documentation updates
- `refactor/description` - Code refactoring
- `test/description` - Test additions/improvements

### Commit Messages

Follow conventional commit format:

```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Formatting, no code change
- `refactor`: Code restructuring
- `test`: Adding tests
- `chore`: Maintenance

Examples:
```
feat(audio): add real-time transcription
fix(backend): resolve database connection timeout
docs(api): update authentication examples
```

### Pull Requests

1. **Small, focused changes**: One feature or fix per PR
2. **Tests required**: All new functionality must include tests
3. **Documentation**: Update relevant docs for API changes
4. **Performance**: Consider performance impact of changes
5. **Security**: Follow security best practices

### PR Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Documentation update
- [ ] Performance improvement
- [ ] Refactoring

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests pass
- [ ] Manual testing completed

## Security
- [ ] No secrets in code
- [ ] Input validation added
- [ ] Authentication/authorization handled

## Performance
- [ ] No performance regression
- [ ] Benchmarks run if applicable

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests pass
```

## Code Style Guidelines

### Frontend (TypeScript/React)

```typescript
// Use functional components with hooks
const AudioRecorder: React.FC<AudioRecorderProps> = ({ onRecordingComplete }) => {
  const [isRecording, setIsRecording] = useState(false);

  // Use descriptive variable names
  const handleStartRecording = useCallback(() => {
    setIsRecording(true);
  }, []);

  return (
    <Button onClick={handleStartRecording}>
      {isRecording ? 'Recording...' : 'Start Recording'}
    </Button>
  );
};
```

### Backend (Go)

```go
// Use clear, descriptive function names
func (s *AudioService) ProcessRecording(ctx context.Context, audioData []byte) (*Transcription, error) {
    // Handle errors explicitly
    if len(audioData) == 0 {
        return nil, ErrEmptyAudioData
    }

    // Use context for cancellation
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    // Process audio...
    transcription, err := s.whisperClient.Transcribe(ctx, audioData)
    if err != nil {
        return nil, fmt.Errorf("transcription failed: %w", err)
    }

    return transcription, nil
}
```

### AI Services (Python)

```python
from typing import Optional, List
import logging

logger = logging.getLogger(__name__)

class TranscriptionService:
    """Service for audio transcription using Whisper.cpp"""

    def __init__(self, model_path: str) -> None:
        self.model_path = model_path
        self._model: Optional[WhisperModel] = None

    async def transcribe(
        self,
        audio_data: bytes,
        language: Optional[str] = None
    ) -> List[TranscriptionSegment]:
        """Transcribe audio data to text segments."""
        try:
            if not self._model:
                await self._load_model()

            # Process transcription...
            segments = await self._process_audio(audio_data, language)
            logger.info(f"Transcribed {len(segments)} segments")

            return segments

        except Exception as e:
            logger.error(f"Transcription failed: {e}")
            raise TranscriptionError(f"Failed to transcribe audio: {e}")
```

## Testing Guidelines

### Unit Tests

- Test each function in isolation
- Mock external dependencies
- Cover edge cases and error conditions
- Maintain >80% code coverage

### Integration Tests

- Test component interactions
- Use test databases and services
- Test API endpoints end-to-end
- Validate data flow between services

### Performance Tests

- Benchmark critical paths
- Test with realistic data volumes
- Monitor memory usage
- Test concurrent operations

## Documentation

### API Documentation

Use OpenAPI/Swagger for REST APIs:

```go
// @Summary      Transcribe audio file
// @Description  Upload and transcribe an audio file
// @Tags         audio
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "Audio file"
// @Success      200   {object}  TranscriptionResponse
// @Failure      400   {object}  ErrorResponse
// @Router       /api/v1/transcribe [post]
func (h *AudioHandler) TranscribeAudio(c *fiber.Ctx) error {
    // Implementation...
}
```

### Code Comments

- Explain **why**, not **what**
- Document complex algorithms
- Add TODO comments for future improvements
- Use godoc format for Go, JSDoc for TypeScript

## Security Guidelines

### Data Handling

- Never log sensitive data
- Encrypt data at rest and in transit
- Validate all user inputs
- Use parameterized queries

### Authentication

- Use strong JWT secrets
- Implement refresh token rotation
- Add rate limiting
- Log security events

### Dependencies

- Keep dependencies updated
- Review security advisories
- Use dependency scanning tools
- Pin dependency versions

## Performance Guidelines

### Database

- Use database indexes appropriately
- Implement query pagination
- Cache frequently accessed data
- Monitor slow queries

### API Design

- Implement request/response caching
- Use compression for large payloads
- Add request timeouts
- Design for horizontal scaling

### Audio Processing

- Stream large audio files
- Use appropriate audio codecs
- Implement background processing
- Monitor memory usage

## Review Process

1. **Automated checks**: All CI checks must pass
2. **Code review**: At least one maintainer approval
3. **Security review**: For changes affecting auth/data handling
4. **Performance review**: For changes affecting critical paths
5. **Documentation review**: For API or architecture changes

## Release Process

### Version Numbering

We use [Semantic Versioning](https://semver.org/):
- **MAJOR**: Breaking changes
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes

### Release Checklist

- [ ] All tests passing
- [ ] Documentation updated
- [ ] Migration scripts tested
- [ ] Security scan completed
- [ ] Performance benchmarks run
- [ ] Release notes prepared

## Community

### Getting Help

- **Documentation**: Check existing docs first
- **Issues**: Search existing issues before creating new ones
- **Discussions**: Use GitHub Discussions for questions
- **Discord**: Join our development Discord for real-time help

### Reporting Issues

Include:
- Steps to reproduce
- Expected vs actual behavior
- Environment details (OS, versions)
- Relevant logs or error messages
- Screenshots if applicable

### Feature Requests

- Check if feature already requested
- Explain the use case and benefit
- Consider implementation complexity
- Provide mockups or examples if helpful

## Recognition

Contributors are recognized in:
- README.md contributors section
- Release notes
- Annual contributor summary
- Discord contributor role

Thank you for contributing to Self! 🚀