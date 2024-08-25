all: image-resize image-resize-rollback path-expand path-flatten rename-date-prefix

image-resize:
	go build -o ../build/image_resize.exe ./main/image-resize/resize.go

image-resize-rollback:
	go build -o ../build/image_resize_rollback.exe ./main/image-resize-rollback/rollback.go

path-expand:
	go build -o ../build/path_expand.exe ./main/path-expand/expand.go

path-flatten:
	go build -o ../build/path_flatten.exe ./main/path-flatten/flatten.go

rename-date-prefix:
	go build -o ../build/rename_date_prefix.exe ./main/rename-date-prefix/main.go