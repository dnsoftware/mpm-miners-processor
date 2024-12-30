PROJECT="miners-processor"

default:
	echo ${PROJECT}

.PHONY: protogen
protogen:
	# здесь /home/dmitry/include/googleapis - путь к официальному репозиторию googleapis
	# который нужно предварительно склонировать командой: git clone https://github.com/googleapis/googleapis.git в эту (или другую) директорию
	# подробности читать тут: https://laradrom.ru/tag/proto/
	protoc --go_out=. --go-grpc_out=. -I.  -I/home/dmitry/include/googleapis proto/miners.proto
