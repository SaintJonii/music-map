# Mapa Musical TDD

**Un laboratorio de TDD para enriquecer y explorar colecciones musicales.**

---

## 🎯 Visión General

Mapa Musical es un proyecto personal diseñado como **laboratorio de Desarrollo Dirigido por Pruebas (TDD)** y espacio de exploración para:

- **Enriquecer** una biblioteca musical personal con datos del audio (características acústicas) y metadatos discográficos
- **Visualizar** las relaciones y patrones dentro de una colección musical de forma interactiva
- **Aprender** TDD en un contexto práctico y real

### Por qué TDD aquí

Este proyecto es deliberadamente un **laboratorio de aprendizaje en TDD**. Cada funcionalidad será desarrollada escribiendo primero las pruebas, lo que permite:
- Entender el comportamiento esperado antes de implementar
- Mantener código limpio y confiable desde el inicio
- Documentar el comportamiento del sistema a través de los tests

---

## 📚 Estado Actual

**Estado**: 🟡 Fase de Setup y Planificación

Actualmente el proyecto se encuentra en:
- ✅ Estructura base del repositorio
- ✅ Definición de objetivos y visión
- ⏳ Configuración del ambiente de desarrollo (próximo)
- ⏳ Primeros tests y arquitectura (próximo)

**No hay funcionalidad implementada aún.** Estamos en la fase de sentar las bases.

---

## 🗓️ Objetivos Iniciales

### Fase 1: Fundamentos
- [ ] Configurar ambiente de desarrollo (Python, pytest, estructura de proyectos)
- [ ] Definir arquitectura básica mediante TDD
- [ ] Crear primeros tests para extracción de características de audio con librosa

### Fase 2: Enriquecimiento de Metadatos
- [ ] Integración exploratoria con **MusicBrainz** para metadatos discográficos
- [ ] Diseñar estrategia de automatización (búsqueda por ISR, título/artista, etc.)
- [ ] Combinar características de audio + metadatos en un modelo unificado

### Fase 3: Visualización Interactiva
- [ ] Crear visualizaciones básicas de la biblioteca
- [ ] Explorar interactividad (mapas 2D/3D, filtrado, navegación)
- [ ] Integrar herramientas de visualización (aún por definir)

---

## 💾 La Biblioteca Existente

Tienes una biblioteca musical personal que queremos usar como base. Esta será:
1. **Enriquecida** con análisis acústico (usando librosa)
2. **Complementada** con metadatos de fuentes externas (MusicBrainz)
3. **Explorada** a través de visualizaciones interactivas

---

## 🤔 Preguntas Abiertas (A Definir)

A medida que avancemos, exploraremos:

- **MusicBrainz**: ¿Qué tan automático puede ser el enriquecimiento? ¿Qué estrategia de búsqueda usamos?
- **Metadatos discográficos**: ¿Cuáles son esenciales? (álbum, artista, año, sello, etc.)
- **Análisis de audio**: ¿Qué características nos interesan realmente?
- **Visualización**: ¿Qué formato es más útil para explorar la colección?
- **Persistencia**: ¿Cómo guardamos y actualizamos los datos enriquecidos?

---

## 🛠️ Stack Tecnológico (Tentativo)

- **Python** 3.8+
- **librosa** - Análisis de audio
- **pytest** + **pytest-cov** - TDD y cobertura
- **MusicBrainz API** - Exploración de metadatos (integración pendiente)
- **Visualización** - A decidir (plotly, matplotlib, web framework?)

---

## 📋 Estructura del Proyecto

```
mapa-musical-tdd/
├── src/
│   └── mapa_musical/
│       └── __init__.py
├── tests/
│   ├── unit/
│   ├── integration/
│   └── conftest.py
├── .github/workflows/        # CI/CD (a configurar)
├── requirements.txt
├── requirements-dev.txt
├── pytest.ini
├── README.md
└── .gitignore
```

---

## 🧪 Cómo Correr Tests (Cuando estén implementados)

```bash
# Instalar dependencias
python -m venv venv
source venv/bin/activate
pip install -r requirements-dev.txt

# Ejecutar pruebas
pytest

# Con reporte de cobertura
pytest --cov=src
```

---

## 📝 Notas Importantes

- **No hay guía de uso aún** - Primero definiremos la arquitectura
- **No hay ejemplos ejecutables** - Se añadirán cuando haya código funcional
- **Los objetivos pueden cambiar** - Este es un proyecto de exploración
- **Enfoque TDD** - El código será guiado por tests desde el principio

---

## 🔗 Recursos

- [librosa Documentation](https://librosa.org/)
- [pytest Documentation](https://docs.pytest.org/)
- [MusicBrainz API](https://musicbrainz.org/doc/Development)
- [Test-Driven Development (Wikipedia)](https://en.wikipedia.org/wiki/Test-driven_development)

---

**Proyecto personal de aprendizaje y exploración. 🎵🧪**
