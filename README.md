# Система проверки на плагиат (Antiplagiat System)

Информационная система для организации хранения студенческих работ и формирования отчетов по результатам проверки на заимствования (антиплагиат).

## Архитектура системы

Система построена на микросервисной архитектуре с четким разделением ответственности между сервисами.

### Компоненты системы

#### 1. API Gateway
- **Назначение**: Центральный сервис-посредник, принимающий все HTTP-запросы от клиентов и маршрутизирующий их к соответствующим микросервисам
- **Технологии**: Go, Chi Router, gRPC клиенты, QuickChart API
- **Порт**: 8080 (по умолчанию)
- **Функции**:
  - Прием HTTP-запросов от клиентов
  - Преобразование HTTP-запросов в gRPC-вызовы
  - Обработка ошибок и преобразование gRPC-ошибок в HTTP-статусы
  - Единая точка входа для всех клиентских запросов
  - Генерация облаков слов для визуализации работ (интеграция с QuickChart API)

#### 2. Storage Service (Сервис хранения файлов)
- **Назначение**: Отвечает за хранение и выдачу файлов
- **Технологии**: Go, gRPC, PostgreSQL, MinIO (S3-совместимое хранилище)
- **Порт**: 5001
- **Функции**:
  - Генерация URL для загрузки файлов в S3
  - Верификация загруженных файлов
  - Генерация временных URL для скачивания файлов
  - Хранение метаданных о файлах в PostgreSQL
  - Получение списка файлов по заданию
- **Хранилища**:
  - **PostgreSQL**: Метаданные о файлах (student_id, task_id, file_id, updated_at, status)
  - **MinIO/S3**: Физическое хранение файлов

#### 3. Plagiarism Service (Сервис анализа на плагиат)
- **Назначение**: Отвечает за проведение анализа, хранение результатов (отчетов) и выдачу отчетов
- **Технологии**: Go, gRPC, PostgreSQL
- **Порт**: 6001
- **Функции**:
  - Проведение анализа на плагиат для всех работ по заданию
  - Попарное сравнение файлов
  - Хранение отчетов о плагиате в PostgreSQL
  - Кэширование результатов анализа
  - Выдача отчетов по заданиям
- **Зависимости**: Использует Storage Service для получения файлов для анализа

### Инфраструктура

- **PostgreSQL** (2 экземпляра):
  - `storage-postgres` (порт 5432): База данных для Storage Service
  - `plagiarism-postgres` (порт 5434): База данных для Plagiarism Service
- **MinIO**: S3-совместимое хранилище для файлов (порты 9000, 9001)
- **Docker Compose**: Оркестрация всех сервисов и зависимостей

## Алгоритм определения плагиата

Система использует алгоритм на основе n-грамм и метрики Jaccard для определения схожести текстов.

### Основные принципы:

1. **Извлечение текста**: Текст извлекается из загруженных файлов (поддерживаются различные форматы)

2. **Предобработка текста**:
   - Приведение к нижнему регистру
   - Нормализация пробелов
   - Очистка от специальных символов

3. **Разбиение на n-граммы**:
   - По умолчанию используется размер n-граммы = 3 (триграммы)
   - Текст разбивается на последовательности из N слов
   - Для коротких текстов используется упрощенный алгоритм сравнения по словам

4. **Вычисление схожести (метрика Jaccard)**:
   ```
   Similarity = |Intersection(n-grams1, n-grams2)| / |Union(n-grams1, n-grams2)|
   ```
   - Где Intersection - количество общих n-грамм
   - Union - общее количество уникальных n-грамм в обоих текстах

5. **Определение плагиата**:
   - Порог плагиата: **0.7 (70%)**
   - Если схожесть >= 0.7, работа считается плагиатом
   - Учитывается время сдачи: если существует более ранняя сдача работы другим студентом, это дополнительный признак плагиата

6. **Попарное сравнение**:
   - Для каждого задания все файлы сравниваются попарно
   - Для каждого студента выбирается отчет с максимальной схожестью

7. **Кэширование результатов**:
   - Результаты анализа сохраняются в базе данных
   - При повторном запросе, если файлы не изменились, возвращаются кэшированные результаты
   - Переанализ выполняется только при появлении новых или измененных файлов

## Пользовательские сценарии и технические сценарии взаимодействия

### Сценарий 1: Загрузка работы студентом

**User Flow:**
1. Студент отправляет запрос на загрузку работы

**Технический сценарий:**

```
Client (HTTP) 
  → API Gateway: POST /api/files
    {
      "task_id": "task_123",
      "student_id": "student_456"
    }
  
API Gateway 
  → Storage Service (gRPC): GenerateUploadURL
    {
      StudentId: "student_456",
      TaskId: "task_123"
    }
  
Storage Service:
  1. Проверяет существование записи в PostgreSQL
  2. Генерирует уникальный file_id
  3. Создает presigned URL для загрузки в MinIO
  4. Сохраняет метаданные в PostgreSQL (status: "pending")
  
Storage Service 
  → API Gateway: GenerateUploadURLResponse
    {
      Url: "https://minio.../presigned-upload-url"
    }
  
API Gateway 
  → Client (HTTP): 200 OK
    {
      "upload_url": "https://minio.../presigned-upload-url"
    }
  
Client:
  → MinIO (Direct): PUT файл по presigned URL
  
Client (HTTP)
  → API Gateway: POST /api/files/verify
    {
      "task_id": "task_123",
      "student_id": "student_456"
    }
  
API Gateway
  → Storage Service (gRPC): VerifyUploadedFile
    {
      StudentId: "student_456",
      TaskId: "task_123"
    }
  
Storage Service:
  1. Проверяет наличие файла в MinIO
  2. Обновляет статус в PostgreSQL (status: "verified")
  3. Обновляет updated_at
  
Storage Service
  → API Gateway: VerifyUploadedFileResponse
    {
      FileId: "file_uuid"
    }
  
API Gateway
  → Client (HTTP): 200 OK
    {
      "file_id": "file_uuid"
    }
```

### Сценарий 2: Запрос анализа на плагиат преподавателем

**User Flow:**
1. Преподаватель запрашивает отчет по проверке работ на плагиат для конкретного задания

**Технический сценарий:**

```
Client (HTTP)
  → API Gateway: POST /api/analysis/task_123
  
API Gateway
  → Plagiarism Service (gRPC): GetPlagiarismReport
    {
      TaskId: "task_123"
    }
  
Plagiarism Service:
  1. Проверяет наличие задачи в своей БД
  2. Если задачи нет - создает новую запись
  
Plagiarism Service
  → Storage Service (gRPC): ListTaskFiles
    {
      TaskId: "task_123"
    }
  
Storage Service:
  1. Запрашивает из PostgreSQL все файлы по task_id
  2. Возвращает список с student_id, updated_at, status
  
Storage Service
  → Plagiarism Service: ListTaskFilesResponse
    {
      Items: [
        {StudentId: "s1", UpdatedAt: "...", Status: "verified"},
        {StudentId: "s2", UpdatedAt: "...", Status: "verified"}
      ]
    }
  
Plagiarism Service:
  1. Проверяет, нужно ли переанализировать (сравнивает updated_at с AnalysisStartedAt)
  2. Если есть кэш и файлы не изменились - возвращает кэшированные результаты
  3. Если нужен новый анализ:
     a. Для каждой пары файлов:
        - Запрашивает download URL у Storage Service
        - Извлекает текст из файлов
        - Сравнивает тексты (n-граммы + Jaccard)
        - Сохраняет отчет в БД
     b. Для каждого студента выбирает отчет с максимальной схожестью
  
Plagiarism Service
  → API Gateway: GetPlagiarismReportResponse
    {
      Reports: [
        {
          Student: "s1",
          StudentWithSimilarFile: "s2",
          MaxSimilarity: 0.85,
          FileHandedOverAt: "..."
        }
      ],
      StartedAt: "..."
    }
  
API Gateway
  → Client (HTTP): 200 OK
    {
      "task_id": "task_123",
      "started_at": "2024-01-01T12:00:00Z",
      "reports": [
        {
          "student": "s1",
          "student_with_similar_file": "s2",
          "max_similarity": 0.85,
          "file_handed_over_at": "2024-01-01T10:00:00Z"
        }
      ]
    }
```

### Сценарий 3: Получение ссылки на скачивание файла

**User Flow:**
1. Преподаватель или студент запрашивает ссылку для скачивания работы

**Технический сценарий:**

```
Client (HTTP)
  → API Gateway: GET /api/files/{task_id}/{student_id}/download
  
API Gateway
  → Storage Service (gRPC): GenerateDownloadURL
    {
      StudentId: "student_456",
      TaskId: "task_123",
      FromInside: false
    }
  
Storage Service:
  1. Проверяет наличие файла в PostgreSQL
  2. Генерирует presigned URL для скачивания из MinIO
  3. URL имеет ограниченное время жизни (по умолчанию 5 минут)
  
Storage Service
  → API Gateway: GenerateDownloadURLResponse
    {
      Url: "https://minio.../presigned-download-url"
    }
  
API Gateway
  → Client (HTTP): 200 OK
    {
      "url": "https://minio.../presigned-download-url"
    }
  
Client:
  → MinIO (Direct): GET файл по presigned URL
```

### Сценарий 4: Генерация облака слов для работы

**User Flow:**
1. Преподаватель или студент запрашивает визуализацию работы в виде облака слов

**Технический сценарий:**

```
Client (HTTP)
  → API Gateway: GET /api/files/{task_id}/{student_id}/wordcloud?format=png
  
API Gateway
  → Storage Service (gRPC): GenerateDownloadURL
    {
      StudentId: "student_456",
      TaskId: "task_123",
      FromInside: true
    }
  
Storage Service
  → API Gateway: GenerateDownloadURLResponse
    {
      Url: "https://minio.../presigned-download-url"
    }
  
API Gateway:
  1. Скачивает файл по presigned URL
  2. Извлекает текст из файла (TextExtractor)
  3. Подготавливает запрос к QuickChart API
  
API Gateway
  → QuickChart API (HTTP POST): https://quickchart.io/wordcloud
    {
      "text": "извлеченный текст из файла...",
      "format": "png",
      "width": 1000,
      "height": 1000,
      "fontScale": 15,
      "scale": "linear",
      "maxNumWords": 200,
      "minWordLength": 3,
      "removeStopwords": true,
      "language": "ru"
    }
  
QuickChart API
  → API Gateway: Image data (PNG/SVG)
  
API Gateway
  → Client (HTTP): 200 OK
    Content-Type: image/png
    [Binary image data]
  
Client:
  Отображает изображение облака слов
```

### Обработка ошибок

Система обрабатывает следующие типы ошибок:

1. **Ошибки валидации** (HTTP 400):
   - Отсутствие обязательных полей (task_id, student_id)
   - Некорректный формат данных

2. **Ошибки "не найдено"** (HTTP 404):
   - Файл не существует
   - Задание не найдено

3. **Ошибки внешних сервисов** (HTTP 502):
   - Один из микросервисов недоступен
   - Ошибка подключения к базе данных
   - Ошибка подключения к MinIO

4. **Внутренние ошибки** (HTTP 500):
   - Непредвиденные ошибки при обработке

При недоступности одного из микросервисов API Gateway корректно обрабатывает ошибку и возвращает соответствующий HTTP-статус клиенту.

## Запуск системы

### Требования
- Docker
- Docker Compose

### Запуск

```bash
docker compose up
```

Система будет доступна по адресу:
- API Gateway: http://localhost:8080
- MinIO Console: http://localhost:9001 (minioadmin/minioadmin)

### Переменные окружения

Основные переменные окружения можно настроить в `docker-compose.yaml`:
- `HTTP_PORT` - порт API Gateway (по умолчанию 8080)
- `STORAGE_ADDR` - адрес Storage Service
- `ANALYSIS_ADDR` - адрес Plagiarism Service

## API Endpoints

### POST /api/files
Генерация URL для загрузки файла

**Request:**
```json
{
  "task_id": "task_123",
  "student_id": "student_456"
}
```

**Response:**
```json
{
  "upload_url": "https://..."
}
```

**Описание:**
- Генерирует presigned URL для загрузки файла в S3 хранилище
- URL действителен ограниченное время (настраивается в Storage Service)
- После получения URL клиент должен выполнить PUT запрос с файлом по этому адресу

### POST /api/files/verify
Верификация загруженного файла

**Request:**
```json
{
  "task_id": "task_123",
  "student_id": "student_456"
}
```

**Response:**
```json
{
  "file_id": "uuid"
}
```

**Описание:**
- Проверяет наличие загруженного файла в хранилище
- Возвращает file_id при успешной верификации
- Если файл не найден, возвращает ошибку 404

### GET /api/files/{task_id}/{student_id}/download
Получение URL для скачивания файла

**Response:**
```json
{
  "url": "https://..."
}
```

### POST /api/analysis/{task_id}
Запрос анализа на плагиат

**Path Parameters:**
- `task_id` - идентификатор задания

**Response:**
```json
{
  "task_id": "task_123",
  "started_at": "2024-01-01T12:00:00Z",
  "reports": [
    {
      "student": "s1",
      "student_with_similar_file": "s2",
      "max_similarity": 0.85,
      "file_handed_over_at": "2024-01-01T10:00:00Z"
    }
  ]
}
```

**Описание:**
- Запускает анализ всех работ по указанному заданию на предмет плагиата
- Сравнивает файлы попарно используя алгоритм n-грамм и метрику Jaccard
- Для каждого студента возвращает отчет с максимальной схожестью
- Результаты кэшируются и пересчитываются только при изменении файлов
- Порог плагиата: 0.7 (70% схожести)

### GET /api/files/{task_id}/{student_id}/wordcloud
Генерация облака слов для присланной работы

**Query Parameters:**
- `format` (опционально): Формат изображения - `png` (по умолчанию) или `svg`

**Response:**
Возвращает изображение облака слов (PNG или SVG) с заголовком `Content-Type: image/png` или `image/svg+xml`.

**Пример использования:**
```
GET /api/files/task_123/student_456/wordcloud?format=png
```

**Особенности:**
- Автоматически извлекает текст из загруженного файла
- Удаляет стоп-слова (для русского языка)
- Показывает наиболее часто встречающиеся слова
- Размер изображения: 1000x1000 пикселей
- Максимум 200 слов в облаке
- Минимальная длина слова: 3 символа

**Технические детали:**
- Использует [QuickChart Word Cloud API](https://quickchart.io/documentation/word-cloud-api/) для генерации визуализации
- Поддерживает форматы PNG и SVG
- Автоматически обрабатывает русский и английский текст

## Тестирование

Для тестирования API используется Postman коллекция:
- Файл: `test/postman/api-gateway.postman_collection.json`
- Базовый URL: `http://localhost:8080`

## Структура проекта

```
Antiplagiat_System/
├── api_gateway/          # API Gateway сервис
├── services/
│   ├── storage_service/  # Сервис хранения файлов
│   └── plagiarism_service/ # Сервис анализа на плагиат
├── test/
│   ├── postman/          # Postman коллекция
│   └── test_files/       # Тестовые файлы
├── docker-compose.yaml   # Конфигурация Docker Compose
└── README.md             # Документация
```

