# Mapa Musical TDD

**Enriquece y visualiza bibliotecas musicales utilizando análisis de audio y visualización de datos.** Este proyecto aprovecha [librosa](https://librosa.org/) para extraer características significativas de archivos de audio y ayudarte a explorar tu colección musical.

Proyecto construido con **Desarrollo Dirigido por Pruebas (TDD)** | Built with **Test-Driven Development (TDD)**

---

## 🎯 Propósito del Proyecto

Mapa Musical está diseñado para resolver el desafío de entender y organizar grandes bibliotecas musicales. Al aplicar técnicas de análisis de audio usando librosa, este proyecto:

- **Enriquece** metadatos musicales con características de audio extraídas (tempo, características espectrales, etc.)
- **Visualiza** relaciones y patrones dentro de tu colección musical
- **Clasifica** pistas basadas en sus propiedades acústicas
- **Permite** la exploración de bibliotecas musicales impulsada por datos

## ✨ Características Principales

- 🎵 **Clasificación Musical** - Categoriza automáticamente pistas basadas en análisis de audio
- 📊 **Extracción de Características de Audio** - Extrae y analiza características como:
  - Tempo y seguimiento de ritmo
  - Análisis espectral (MFCCs, características de cromaticidad, etc.)
  - Volumen y rango dinámico
  - Huella digital de audio
- 🗺️ **Visualización de Biblioteca** - Crea visualizaciones interactivas de tu colección musical
- 🧪 **Cobertura de Pruebas del 100%** - Cada característica está respaldada por un conjunto de pruebas integral
- 📈 **Información Impulsada por Datos** - Comprende patrones en tu gusto musical y colección

## 📋 Requisitos

- **Python**: 3.8 o superior
- **Dependencias Principales**:
  - `librosa` - Biblioteca de análisis de audio
  - `numpy` - Computación numérica
  - `scipy` - Computación científica
  - `matplotlib` / `plotly` - Visualización de datos (recomendado)
- **Desarrollo**:
  - `pytest` - Marco de pruebas
  - `pytest-cov` - Reporte de cobertura

## 🚀 Instalación

### Clonar el repositorio

```bash
git clone https://github.com/aelaredo/mapa-musical-tdd.git
cd mapa-musical-tdd
```

### Configurar entorno virtual

```bash
python -m venv venv
source venv/bin/activate  # En Windows: venv\Scripts\activate
```

### Instalar dependencias

```bash
pip install -r requirements.txt
```

### Instalar dependencias de desarrollo (opcional)

```bash
pip install -r requirements-dev.txt
```

## 💻 Ejemplo de Inicio Rápido

```python
from mapa_musical import MusicAnalyzer

# Inicializa el analizador
analyzer = MusicAnalyzer()

# Carga un archivo de audio
features = analyzer.extract_features('path/to/audio/file.mp3')

# Analiza los resultados
print(f"Tempo: {features['tempo']} BPM")
print(f"Indicadores de género: {features['genre_features']}")

# Clasifica la pista
classification = analyzer.classify(features)
print(f"Clasificación predicha: {classification}")
```

## 🧪 Pruebas y Desarrollo Dirigido por Pruebas

Este proyecto mantiene una **cobertura de pruebas del 100%** mediante prácticas estrictas de TDD. Todo el código se escribe primero las pruebas.

### Ejecutar pruebas

```bash
pytest
```

### Ejecutar pruebas con reporte de cobertura

```bash
pytest --cov=src --cov-report=html
```

### Ejecutar pruebas en modo vigilancia (opcional)

```bash
pytest-watch
```

## 📚 Guía de Uso

### 1. Extracción Básica de Características de Audio

Extrae características de un único archivo de audio para entender sus características:

```python
analyzer = MusicAnalyzer()
features = analyzer.extract_features('song.mp3')
```

### 2. Procesamiento por Lotes de una Biblioteca Musical

Procesa una carpeta completa para enriquecer tu colección musical:

```python
analyzer = MusicAnalyzer()
library_features = analyzer.process_directory('path/to/music/library/')
```

### 3. Visualiza Tu Colección Musical

Crea visualizaciones para explorar relaciones:

```python
from mapa_musical import MusicVisualizer

visualizer = MusicVisualizer(library_features)
visualizer.plot_2d_map()  # Gráfico de dispersión 2D de tu biblioteca
visualizer.plot_tempo_distribution()
visualizer.plot_genre_clustering()
```

## 📊 Estructura del Proyecto

```
mapa-musical-tdd/
├── src/
│   ├── mapa_musical/
│   │   ├── __init__.py
│   │   ├── analyzer.py          # Análisis de audio principal
│   │   ├── classifier.py        # Clasificación musical
│   │   ├── visualizer.py        # Utilidades de visualización
│   │   └── features.py          # Extracción de características
│   └── ...
├── tests/
│   ├── unit/                    # Pruebas unitarias
│   ├── integration/             # Pruebas de integración
│   └── conftest.py              # Accesorios de pytest
├── .github/workflows/           # Pipelines de CI/CD
├── requirements.txt             # Dependencias de producción
├── requirements-dev.txt         # Dependencias de desarrollo
└── README.md
```

## 🔄 Estado del Proyecto

**Estado**: 🟡 **Desarrollo Activo**

Este proyecto está en desarrollo activo. Las características principales se están construyendo siguiendo principios de TDD. Espera actualizaciones y mejoras regulares.

### Características Planeadas

- [ ] Análisis de secuencia de audio en tiempo real
- [ ] Algoritmos de agrupamiento avanzados
- [ ] Panel de visualización basado en web
- [ ] Servidor de API para integraciones externas
- [ ] Soporte para múltiples formatos de audio
- [ ] Clasificación de género basada en aprendizaje automático

### Limitaciones Conocidas

- Actualmente procesa archivos de audio sin conexión (soporte de transmisión próximamente)
- La aceleración de GPU aún no está implementada (puede ser lento para bibliotecas muy grandes)

## 🤝 Contribuciones

¡Bienvenidas las contribuciones! Aquí te mostramos cómo comenzar:

1. **Haz un fork** del repositorio
2. **Crea una rama de característica** (`git checkout -b feature/amazing-feature`)
3. **Escribe primero las pruebas** (siguiendo principios de TDD)
4. **Implementa tu característica** con cobertura de pruebas del 100%
5. **Confirma tus cambios** (`git commit -m 'Add amazing feature'`)
6. **Sube a tu fork** (`git push origin feature/amazing-feature`)
7. **Abre una solicitud de extracción** y describe tus cambios

### Directrices de Contribución

- Siempre sigue TDD: escribe pruebas antes de la implementación
- Mantén cobertura de pruebas del 100%
- Incluye docstrings para todas las funciones y clases
- Sigue la guía de estilo PEP 8
- Asegúrate de que CI/CD pase antes de enviar el PR

## 📝 Licencia

Sin licencia asignada aún. Por favor, consulta el archivo LICENSE para más detalles.

## 🔗 Enlaces y Recursos

- [Documentación de librosa](https://librosa.org/)
- [Documentación de pytest](https://docs.pytest.org/)
- [Mejores Prácticas de Desarrollo Dirigido por Pruebas](https://en.wikipedia.org/wiki/Test-driven_development)

## 📞 Soporte y Contacto

### Obtener Ayuda

- 🐛 **Reportes de Errores**: Abre un [issue](https://github.com/aelaredo/mapa-musical-tdd/issues)
- 💡 **Solicitudes de Características**: Abre un [issue](https://github.com/aelaredo/mapa-musical-tdd/issues)
- 📧 **Contacto Directo**: Ponte en contacto con los mantenedores

### Problemas de GitHub

Para reportes de errores y solicitudes de características, por favor utiliza la página de [Problemas de GitHub](https://github.com/aelaredo/mapa-musical-tdd/issues).

---

**Hecho con ❤️ para amantes de la música y entusiastas de datos**
