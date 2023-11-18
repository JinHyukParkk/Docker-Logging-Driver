# Docker Logging Driver
-------------
## 1. 개요
효율적 로그 관리를 위한 Docker Logging Driver 입니다.  
추가된 기능은 다음과 같습니다.
- 기존 Docker Logging Driver에서 실행과 동시에 Log 데이터 저장 공간을 Mount하여 별도로 공간 확보
- 하나의 파일에 Log 데이터가 쌓이면서 일정 크기가 되었을 때 새로운 파일로 Lotation

## 2. 구성 환경 
* OS : CentOs7
* Language : Golang
* Docker daemon server 
* Docker Private Registry - container
* Docker Logging Driver

## 3. 실행 방법
```
# make 
```



