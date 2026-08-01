# 🎵 Mapa Musical — Guía de Onboarding

> *"Tu colección de música, mapeada como un universo personal."*

Bienvenido a **Mapa Musical**. Este documento te explica todo lo que necesitás saber sobre el proyecto: qué hace, cómo funciona, y el *por qué* detrás de cada decisión. Está escrito para que lo entienda tanto un ingeniero de software como alguien que nunca programó — si algo no queda claro, el [glosario](#-glosario) está al final para rescatarte.

---

## 🗺️ ¿Qué es Mapa Musical?

Imaginá un mapa estelar, pero en vez de estrellas, cada punto es **una canción tuya**. Las canciones que suenan parecido están cerca; las que no tienen nada que ver, lejos. Podés colorear por género, filtrar por década, hacer clic en un punto y escuchar la canción.

La inspiración directa es [**Every Noise at Once**](https://everynoise.com/), un mapa interactivo de 6.000+ géneros musicales creado por Glenn McDonald (ex-Spotify). La diferencia clave: Every Noise mapea géneros globales con datos de millones de usuarios. **Mapa Musical mapea TU biblioteca personal**, con los archivos que tenés en tu disco.

---

## 🎧 El Universo del Audio (sin sufrir)

Si venís del mundo del software pero nunca tocaste audio, esta sección es para vos.

### ¿Qué hay dentro de un archivo de audio?

Un archivo como `cancion.mp3` o `track.wav` contiene dos cosas:

| Capa | Qué es | Ejemplo |
|---|---|---|
| **Audio crudo** | Una lista enorme de números que representan cómo se mueve el aire (la onda sonora) | `[0.0, 0.15, 0.31, 0.45, ...]` (44.100 números por segundo) |
| **Metadatos** | Etiquetas con información: título, artista, álbum, género, año, etc. | `TITLE=Bohemian Rhapsody`, `ARTIST=Queen`, `BPM=72` |

### Los tres formatos que soportamos

| Formato | Compresión | Calidad | ¿Por qué lo soportamos? |
|---|---|---|---|
| **WAV** | Ninguna (lossless) | La original | Es el más simple de leer. Ideal para empezar. |
| **MP3** | Con pérdida (lossy) | Buena, ocupa ~10x menos | Es el formato más común del mundo. |
| **FLAC** | Sin pérdida (lossless) | Idéntica al original, ocupa ~50% menos que WAV | Estándar en colecciones de audiófilos. |

> 💡 **Lossless vs Lossy**: *Lossless* = no se pierde nada, como un ZIP. *Lossy* = se descarta información que el oído humano apenas nota, como un JPEG de una foto. FLAC es lossless, MP3 es lossy.

### PCM: la materia prima

Independientemente del formato, internamente todo se convierte a **PCM** (*Pulse-Code Modulation*). Es simplemente una lista de números (samples) que miden la amplitud de la onda sonora en intervalos regulares:

```
Sample rate: 44.100 veces por segundo (44.1 kHz)
Bit depth:   16 bits por sample (65.536 niveles posibles)
Canales:     1 (mono) o 2 (estéreo)
```

Un archivo de 3 minutos en calidad CD (44.1 kHz, 16-bit, estéreo) tiene:
- 44.100 × 180 segundos = **~7.9 millones de samples por canal**
- 2 canales = **~15.8 millones de samples totales**

¡Por eso necesitamos algoritmos eficientes!

---

## 🔬 Features Acústicas: ¿qué medimos y por qué?

Extraemos 9 características de cada canción. Cada una captura un aspecto distinto de "cómo suena". Acá va una explicación sin matemática innecesaria:

| Feature | ¿Qué mide? | Intuición humana | Rango |
|---|---|---|---|
| **RMS Energy** | Volumen/intensidad promedio | Qué tan "fuerte" suena | 0 a 1 |
| **ZCR** (Zero-Crossing Rate) | Qué tan rápido cambia la onda de positivo a negativo | Sonidos agudos/ruidosos tienen ZCR alto; graves/bajos tienen ZCR bajo | 0 a ~4000 |
| **Spectral Centroid** | El "centro de gravedad" de las frecuencias | Un violín tiene centroide alto (brillante); un contrabajo tiene centroide bajo (oscuro) | Hz |
| **BPM** (Beats Per Minute) | Tempo / velocidad | 60 BPM = lento (balada), 120 BPM = bailable, 180 BPM = frenético | ~60–200 |
| **Chroma** | Qué notas musicales están presentes | Un acorde de Do mayor activa los bins C, E, G | 12 valores (uno por nota) |
| **MFCCs** (Mel-Frequency Cepstral Coefficients) | El "timbre" o "color" del sonido | Lo que hace que un piano y una guitarra tocando la misma nota suenen diferente | 13 números |
| **Key** | Tonalidad (escala musical) | "Esta canción está en La menor" | Ej: "C major", "A minor" |
| **Danceability** | Qué tan "bailable" es | Combina BPM + energía + ritmo en un solo número | 0 a 1 |
| **Acousticness** | Qué tan "acústica" (vs electrónica) suena | Guitarra criolla = alta. Sintetizador = baja | 0 a 1 |

### ¿Cómo pasamos de "números de aire" a "esta canción es bailable"?

```
Archivo de audio
    │
    ▼
Decoder (WAV/MP3/FLAC) → lista de samples PCM
    │
    ▼
FFT (Fast Fourier Transform) → convierte los samples en frecuencias
    │                              (como pasar de una receta a los ingredientes)
    ▼
Extractores individuales → cada uno aplica su fórmula
    │                        (BPM busca patrones rítmicos,
    │                         Chroma agrupa frecuencias en notas,
    │                         MFCCs modela cómo el oído humano percibe el timbre)
    ▼
TrackFeatures → struct con los 9 valores
```

> 💡 **FFT** es el "traductor" mágico: toma una señal en el tiempo (samples) y la convierte en frecuencias. Es como desarmar un smoothie para saber qué frutas tiene. Sin FFT no podríamos extraer ninguna feature espectral.

---

## 🏷️ Metadata: lo que el archivo ya sabe

Antes de gastar recursos en análisis acústico, **primero leemos lo que el archivo ya dice de sí mismo**. Muchos archivos de audio tienen metadatos incrustados:

| Formato | Sistema de tags | Campos típicos |
|---|---|---|
| MP3 | **ID3v2** | Título, Artista, Álbum, Género, Año, BPM, ISRC, MusicBrainz ID |
| FLAC | **Vorbis Comments** | Lo mismo, pero con nombres de campo diferentes |
| WAV | RIFF chunks (a veces) | Generalmente sin tags ricos |

### Estrategia "tags-first" 🥇

Nuestro pipeline sigue este orden de prioridad para obtener datos:

```
1. Metadatos embebidos (ID3v2 / Vorbis)  ← O(1), instantáneo
2. MusicBrainz API (si hay ISRC o MBID)   ← O(1) + latencia de red
3. Análisis acústico (nuestros extractores) ← O(n) por segundo de audio, lento
```

El 80% de las canciones ya tienen título, artista, álbum, género y hasta BPM en sus tags. Solo extraemos features acústicas cuando los tags no alcanzan.

### ¿Qué es MusicBrainz?

[MusicBrainz](https://musicbrainz.org/) es una "Wikipedia de la música": una base de datos abierta con información de millones de canciones, álbumes y artistas. Si un archivo tiene un **ISRC** (código único internacional de grabación) o un **MBID** (MusicBrainz ID), podemos consultar la API para obtener datos faltantes como fecha de lanzamiento exacta, sello discográfico, o carátula del álbum.

---

## 🏗️ Arquitectura del Proyecto

### Estructura de paquetes

```
mapa-musical-tdd/
│
├── model/          🧬  Tipos de dominio (Track, TrackFeatures, Collection)
│   └── track.go        Structs puros, cero dependencias externas
│
├── audio/          🎧  Todo lo relacionado con procesamiento de audio
│   ├── decoder.go      Interfaz Decoder + detector de formato por magic bytes
│   ├── wav.go          Decodificador WAV (16/24/32-bit)
│   ├── mp3.go          Decodificador MP3 (MPEG Layer III)
│   ├── flac.go         Decodificador FLAC (lossless)
│   ├── features.go     Interfaz FeatureExtractor + 9 extractores
│   └── dsp.go          Primitivas DSP: FFT, ventanas, mel scale, autocorrelación
│
├── metadata/       🏷️  Lectura de tags y enriquecimiento externo
│   ├── reader.go       TagReader: extrae ID3v2 / Vorbis comments
│   └── enricher.go     MusicBrainzClient: consulta la API de MusicBrainz
│
├── storage/        💾  Persistencia de resultados
│   └── repository.go   SQLite: guarda tracks + features, consulta por ID
│
├── cmd/mapa-musical/  🖥️  CLI: punto de entrada
│   └── main.go         Pipeline completo: decodificar → extraer → tags → persistir
│
├── testdata/       🧪  Fixtures de audio para tests
│   ├── wav/            WAVs sintéticos en 16/24/32-bit
│   ├── mp3/            MP3 de prueba + corrupto
│   └── flac/           FLAC 16-bit lossless
│
└── openspec/       📋  Artefactos SDD (propuesta, specs, diseño, tareas)
    └── specs/          6 especificaciones formales (fuente de verdad)
```

### Data Flow (el camino de una canción)

```
🎵 archivo.mp3
    │
    ├──→ TagReader ──→ Track { título, artista, ISRC, género... }
    │                        │
    │                        ├──→ (si tiene ISRC) MusicBrainzClient ──→ datos extra
    │                        │
    ├──→ DetectFormat ──→ Decoder ──→ []float64 samples PCM
    │                                        │
    │                                        ▼
    │                              FeatureExtractor
    │                                   │
    │                    ┌──────────────┼──────────────┐
    │                    ▼              ▼              ▼
    │                  RMS           Chroma          BPM
    │                  ZCR           MFCCs           Key
    │                  Centroid      Danceability    Acousticness
    │                                   │
    │                                   ▼
    │                            TrackFeatures
    │                                   │
    └───────────────────────────────────┘
                                        │
                                        ▼
                              Repository.Save(track, features)
                                        │
                                        ▼
                                    SQLite 💾
```

### Principios de diseño

| Principio | Cómo se aplica |
|---|---|
| **TDD estricto** | Todo test se escribe PRIMERO (RED), falla, y recién después se implementa (GREEN). Cobertura ≥ 80%. |
| **Interfaces sobre implementaciones** | Cada frontera entre paquetes es una interfaz (`Decoder`, `FeatureExtractor`, `Repository`). Fácil de testear con mocks. |
| **Constructor injection** | Las dependencias se pasan explícitamente en el constructor. Cero magia, cero globales. |
| **Tags-first** | Leer metadatos existentes antes de gastar CPU en análisis acústico. Más rápido, más escalable. |
| **Flat package layout** | Paquetes por dominio (`model`, `audio`, `metadata`, `storage`), no por tipo de archivo. Simple, refactorizable. |

---

## 🧪 Cómo trabajamos: SDD + TDD

Usamos **SDD** (*Spec-Driven Development*): antes de escribir una línea de código, definimos:

1. **Propuesta** — ¿qué problema resolvemos? ¿qué entregamos?
2. **Especificaciones** — requisitos formales con escenarios Given/When/Then
3. **Diseño** — decisiones de arquitectura, interfaces, data flow
4. **Tareas** — pasos concretos de implementación, ordenados por dependencia

Y dentro de cada tarea, aplicamos **TDD** (*Test-Driven Development*):

```
🔴 RED     → Escribir un test que falle
🟢 GREEN   → Escribir el mínimo código para que pase
🔵 REFACTOR → Limpiar el código sin cambiar comportamiento
```

### Comandos útiles

```bash
go test ./...           # Correr todos los tests
go test ./... -cover    # Con reporte de cobertura
go test ./audio/... -v  # Tests de audio con output detallado
go vet ./...            # Análisis estático
go build ./...          # Verificar que todo compile
```

---

## 📊 Estado Actual del Proyecto

| Métrica | Valor |
|---|---|
| **Lenguaje** | Go 1.26.5 |
| **Tests** | ~50 pasando |
| **Cobertura** | 83.1% global |
| **Modelo** | 100% |
| **Audio (decoders + extractores)** | 88.7% |
| **Metadata** | 89.4% |
| **Storage** | 82.4% |
| **TDD** | Estricto (RED → GREEN → REFACTOR) |
| **SDD** | Ciclo completo archivado |

### Lo que ya funciona

- ✅ Decodificación de WAV (16/24/32-bit), MP3, FLAC
- ✅ Detección automática de formato por magic bytes
- ✅ 9 extractores de features acústicas (RMS, ZCR, Centroid, BPM, Chroma, MFCCs, Key, Danceability, Acousticness)
- ✅ Lectura de tags ID3v2 y Vorbis Comments
- ✅ Enriquecimiento vía MusicBrainz API (por ISRC o MBID)
- ✅ Persistencia en SQLite
- ✅ Pipeline CLI completo: `go run ./cmd/mapa-musical/ cancion.mp3`
- ✅ Code review aprobado (10 issues corregidos)

### Lo que viene después (post-MVP)

- 🔲 Integración con **Navidrome/Subsonic API** (biblioteca centralizada, acceso remoto)
- 🔲 Reducción dimensional con **PCA/UMAP** (colapsar 9 features en coordenadas 2D para el mapa)
- 🔲 **Mapa interactivo** (frontend web con D3.js/Three.js, click→reproducir, filtros por género/año/BPM)
- 🔲 Sincronización automática con carpeta de música

---

## 📚 Glosario

### 🎧 Términos de Audio

| Término | Definición |
|---|---|
| **Sample** | Un valor numérico que representa la amplitud de una onda sonora en un instante. |
| **Sample Rate** | Cuántos samples por segundo. 44.100 Hz = calidad CD. |
| **Bit Depth** | Cuánta precisión tiene cada sample. 16-bit = 65.536 niveles. 24-bit = 16.7 millones. |
| **PCM** (Pulse-Code Modulation) | Representación digital de audio: una secuencia de samples. |
| **Mono / Estéreo** | 1 canal vs 2 canales (izquierdo y derecho). |
| **WAV** | Formato de audio sin compresión. Archivos grandes, calidad perfecta. |
| **MP3** | Formato con compresión lossy. El más común. Reduce tamaño ~10x descartando frecuencias poco audibles. |
| **FLAC** | Formato lossless comprimido. Calidad idéntica a WAV pero ~50% más chico. |
| **ID3v2** | Sistema de metadatos para MP3. Permite guardar título, artista, álbum, género, BPM, ISRC, etc. |
| **Vorbis Comments** | Equivalente a ID3v2 pero para FLAC. Mismos datos, distintos nombres de campo. |
| **ISRC** | Código único internacional que identifica una grabación específica (ej: `GBUM71029604`). |
| **MBID** | MusicBrainz Identifier. UUID que identifica un artista, álbum o grabación en MusicBrainz. |
| **FFT** (Fast Fourier Transform) | Algoritmo que convierte una señal de tiempo a frecuencia. El "traductor" fundamental del DSP. |
| **BPM** (Beats Per Minute) | Tempo. 60 = lento, 120 = estándar bailable, 180 = rápido. |
| **RMS** (Root Mean Square) | Medida de energía/intensidad promedio de una señal. |

### 🏗️ Términos de Ingeniería

| Término | Definición |
|---|---|
| **TDD** (Test-Driven Development) | Escribir tests antes que el código de producción. Ciclo RED → GREEN → REFACTOR. |
| **SDD** (Spec-Driven Development) | Planificar con propuesta → especificaciones → diseño → tareas antes de implementar. |
| **SOLID** | Cinco principios de diseño orientado a objetos. En Go aplicamos especialmente ISP (interfaces pequeñas) y DIP (depender de abstracciones). |
| **ISP** (Interface Segregation Principle) | Interfaces chicas y específicas en vez de una interfaz monolítica gigante. |
| **Constructor Injection** | Pasar dependencias por el constructor en vez de usar globales o service locators. |
| **Sentinel Error** | Un error predefinido como variable exportada (`var ErrNotFound = errors.New(...)`) que permite comparación con `errors.Is()`. |
| **Mock** | Objeto falso que simula una dependencia externa en tests (ej: simular una API HTTP sin conectarse a internet). |
| **Golden File / Fixture** | Archivo de prueba pre-grabado que representa una entrada o salida esperada. |
| **Pipeline** | Secuencia de etapas de procesamiento que se ejecutan en orden. |
| **DSP** (Digital Signal Processing) | Procesamiento digital de señales. La base matemática de todo el análisis de audio. |
| **PCA** (Principal Component Analysis) | Técnica para reducir dimensiones. Toma 9 features y las colapsa en 2 para graficar. |
| **UMAP** (Uniform Manifold Approximation and Projection) | Alternativa moderna a PCA. Mejor preservando la estructura local de los datos. |

### 🗺️ Términos del Proyecto

| Término | Definición |
|---|---|
| **Tags-first** | Estrategia de priorizar metadatos existentes sobre análisis acústico. Instantáneo vs lento. |
| **LibrarySource** | Interfaz que abstrae de dónde vienen las canciones. Hoy: filesystem. Mañana: Navidrome API. |
| **Magic Bytes** | Los primeros bytes de un archivo que identifican su formato. WAV empieza con `RIFF`, FLAC con `fLaC`, MP3 con `0xFF 0xFB`. |
| **MusicBrainz** | Base de datos colaborativa de metadatos musicales. La "Wikipedia de la música". |
| **Navidrome** | Servidor de música auto-hosteado compatible con la API Subsonic. |
| **Subsonic API** | Protocolo estándar para servidores de música (Navidrome, Airsonic, Funkwhale). |
| **Every Noise at Once** | Mapa interactivo de géneros musicales. Inspiración visual del proyecto. |
| **OpenSpec** | Formato de artefactos SDD usado en este proyecto (propuesta, specs, diseño, tareas). |
| **Engram** | Sistema de memoria persistente para agentes de IA. Guarda contexto entre sesiones. |

---

## 🔗 Enlaces Útiles

| Recurso | URL |
|---|---|
| Every Noise at Once | https://everynoise.com/ |
| MusicBrainz | https://musicbrainz.org/ |
| Navidrome | https://www.navidrome.org/ |
| OpenSubsonic API docs | https://opensubsonic.netlify.app/ |
| go-dsp (FFT library) | https://github.com/madelynnblue/go-dsp |
| go-audio (WAV) | https://github.com/go-audio/wav |
| dhowden/tag (ID3/Vorbis) | https://github.com/dhowden/tag |
| modernc.org/sqlite | https://pkg.go.dev/modernc.org/sqlite |
| Repositorio del proyecto | https://github.com/SaintJonii/music-map |

---

*Documento generado el 1 de agosto de 2026. Época de Go 1.26.5, 50 tests, 83.1% de cobertura, y cero conocimientos previos de audio requeridos.* 🎧
