module api_gateway

go 1.25

require (
	github.com/Nikita-Smirnov-idk/plagiarism-service v0.0.0-00010101000000-000000000000
	github.com/Nikita-Smirnov-idk/storage-service v0.0.0-00010101000000-000000000000
	github.com/go-chi/chi/v5 v5.2.1
	github.com/ilyakaznacheev/cleanenv v1.5.0
	google.golang.org/grpc v1.77.0
)

require (
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251022142026-3a174f9686a8 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	olympos.io/encoding/edn v0.0.0-20201019073823-d3554ca0b0a3 // indirect
)

replace github.com/Nikita-Smirnov-idk/storage-service => ../services/storage_service

replace github.com/Nikita-Smirnov-idk/plagiarism-service => ../services/plagiarism_service
