ISU01_CUSTOM_DATA = isu01CustomData.sh
ISU02_CUSTOM_DATA = isu02CustomData.sh
SRC_FILE = azuredeploy.json.in
TARGET = azuredeploy.json

${TARGET}: ${SRC_FILE} ${ISU01_CUSTOM_DATA} ${ISU02_CUSTOM_DATA}
	sed \
    -e "s/@isu01CustomData@/`cat ${ISU01_CUSTOM_DATA} | openssl enc -base64 | tr -d '\n'`/g" \
    -e "s/@isu02CustomData@/`cat ${ISU02_CUSTOM_DATA} | openssl enc -base64 | tr -d '\n'`/g" \
    ${SRC_FILE} > ${TARGET}

all: ${TARGET}
