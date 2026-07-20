# Mapa Musical TDD

**Enrich and visualize music libraries using audio analysis and data visualization.** This project leverages [librosa](https://librosa.org/) to extract meaningful features from audio files and helps you organize, classify, and explore your music collection.

Proyecto construido con **Test-Driven Development (TDD)** | Built with **Test-Driven Development (TDD)**

---

## 🎯 Project Purpose

Mapa Musical is designed to solve the challenge of understanding and organizing large music libraries. By applying audio analysis techniques using librosa, this project:

- **Enriches** music metadata with extracted audio features (tempo, spectral characteristics, etc.)
- **Visualizes** relationships and patterns within your music collection
- **Classifies** tracks based on their acoustic properties
- **Enables** data-driven exploration of music libraries

## ✨ Key Features

- 🎵 **Music Classification** - Automatically categorize tracks based on audio analysis
- 📊 **Audio Feature Extraction** - Extract and analyze features like:
  - Tempo and beat tracking
  - Spectral analysis (MFCCs, chroma features, etc.)
  - Loudness and dynamic range
  - Audio fingerprinting
- 🗺️ **Library Visualization** - Create interactive visualizations of your music collection
- 🧪 **100% Test Coverage** - Every feature is backed by comprehensive test suite
- 📈 **Data-Driven Insights** - Understand patterns in your music taste and collection

## 📋 Requirements

- **Python**: 3.8 or higher
- **Core Dependencies**:
  - `librosa` - Audio analysis library
  - `numpy` - Numerical computing
  - `scipy` - Scientific computing
  - `matplotlib` / `plotly` - Data visualization (recommended)
- **Development**:
  - `pytest` - Testing framework
  - `pytest-cov` - Coverage reporting

## 🚀 Installation

### Clone the repository

```bash
git clone https://github.com/aelaredo/mapa-musical-tdd.git
cd mapa-musical-tdd
```

### Set up virtual environment

```bash
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
```

### Install dependencies

```bash
pip install -r requirements.txt
```

### Install development dependencies (optional)

```bash
pip install -r requirements-dev.txt
```

## 💻 Quick Start Example

```python
from mapa_musical import MusicAnalyzer

# Initialize the analyzer
analyzer = MusicAnalyzer()

# Load an audio file
features = analyzer.extract_features('path/to/audio/file.mp3')

# Analyze the results
print(f"Tempo: {features['tempo']} BPM")
print(f"Genre indicators: {features['genre_features']}")

# Classify the track
classification = analyzer.classify(features)
print(f"Predicted classification: {classification}")
```

## 🧪 Testing & Test-Driven Development

This project maintains **100% test coverage** through strict TDD practices. All code is written tests-first.

### Run tests

```bash
pytest
```

### Run tests with coverage report

```bash
pytest --cov=src --cov-report=html
```

### Run tests in watch mode (optional)

```bash
pytest-watch
```

## 📚 Usage Guide

### 1. Basic Audio Feature Extraction

Extract features from a single audio file to understand its characteristics:

```python
analyzer = MusicAnalyzer()
features = analyzer.extract_features('song.mp3')
```

### 2. Batch Processing a Music Library

Process an entire folder to enrich your music collection:

```python
analyzer = MusicAnalyzer()
library_features = analyzer.process_directory('path/to/music/library/')
```

### 3. Visualize Your Music Collection

Create visualizations to explore relationships:

```python
from mapa_musical import MusicVisualizer

visualizer = MusicVisualizer(library_features)
visualizer.plot_2d_map()  # 2D scatter plot of your library
visualizer.plot_tempo_distribution()
visualizer.plot_genre_clustering()
```

## 📊 Project Structure

```
mapa-musical-tdd/
├── src/
│   ├── mapa_musical/
│   │   ├── __init__.py
│   │   ├── analyzer.py          # Core audio analysis
│   │   ├── classifier.py        # Music classification
│   │   ├── visualizer.py        # Visualization utilities
│   │   └── features.py          # Feature extraction
│   └── ...
├── tests/
│   ├── unit/                    # Unit tests
│   ├── integration/             # Integration tests
│   └── conftest.py              # Pytest fixtures
├── .github/workflows/           # CI/CD pipelines
├── requirements.txt             # Production dependencies
├── requirements-dev.txt         # Development dependencies
└── README.md
```

## 🔄 Project Status

**Status**: 🟡 **Active Development**

This project is in active development. Core features are being built following TDD principles. Expect regular updates and improvements.

### Planned Features

- [ ] Real-time audio stream analysis
- [ ] Advanced clustering algorithms
- [ ] Web-based visualization dashboard
- [ ] API server for external integrations
- [ ] Support for multiple audio formats
- [ ] Machine learning-based genre classification

### Known Limitations

- Currently processes offline audio files (streaming support coming soon)
- GPU acceleration not yet implemented (can be slow for very large libraries)

## 🤝 Contributing

We welcome contributions! Here's how to get started:

1. **Fork** the repository
2. **Create a feature branch** (`git checkout -b feature/amazing-feature`)
3. **Write tests first** (following TDD principles)
4. **Implement your feature** with 100% test coverage
5. **Commit your changes** (`git commit -m 'Add amazing feature'`)
6. **Push to your fork** (`git push origin feature/amazing-feature`)
7. **Open a Pull Request** and describe your changes

### Contribution Guidelines

- Always follow TDD: write tests before implementation
- Maintain 100% test coverage
- Include docstrings for all functions and classes
- Follow PEP 8 style guide
- Make sure CI/CD passes before submitting PR

## 📝 License

No license assigned yet. Please see LICENSE file for details.

## 🔗 Links & Resources

- [librosa Documentation](https://librosa.org/)
- [pytest Documentation](https://docs.pytest.org/)
- [Test-Driven Development Best Practices](https://en.wikipedia.org/wiki/Test-driven_development)

## 📞 Support & Contact

### Getting Help

- 🐛 **Bug Reports**: Open an [issue](https://github.com/aelaredo/mapa-musical-tdd/issues)
- 💡 **Feature Requests**: Open an [issue](https://github.com/aelaredo/mapa-musical-tdd/issues)
- 📧 **Direct Contact**: Reach out to the maintainers

### GitHub Issues

For bug reports and feature requests, please use the [GitHub Issues](https://github.com/aelaredo/mapa-musical-tdd/issues) page.

---

**Made with ❤️ for music lovers and data enthusiasts**
