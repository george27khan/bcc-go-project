# bcc-go-project
Итоговый тестовый проект по курсу Go в BCC


oapi-codegen --package=api --generate chi-server api/openapi.yaml > api/server.gen.go

oapi-codegen --config=internal/api/oapi-codegen.yaml internal/api/openapi.yaml
