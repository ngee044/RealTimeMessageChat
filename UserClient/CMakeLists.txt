cmake_minimum_required(VERSION 3.18)

set(PROGRAM_NAME UserClient)
set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED TRUE)

set(SOURCE_FILES
	main.cpp
	Configurations.cpp
	UserClient.cpp
)

set (HEADER_FILES
	Configurations.h
	UserClient.h
)

project(${PROGRAM_NAME} VERSION 1.0.0.0)

add_executable(${PROGRAM_NAME} ${HEADER_FILES} ${SOURCE_FILES})

target_link_libraries(${PROGRAM_NAME} PUBLIC CommonModule Utilities Thread Network)
target_include_directories(${PROGRAM_NAME} PRIVATE "${CMAKE_CURRENT_SOURCE_DIR}")

set(JSON_FILES
	user_client_configurations.json
)

foreach(JSON_FILE IN LISTS JSON_FILES)
	add_custom_command(
		TARGET UserClient POST_BUILD
		COMMAND ${CMAKE_COMMAND} -E copy_if_different
			${CMAKE_CURRENT_SOURCE_DIR}/${JSON_FILE}
			${CMAKE_RUNTIME_OUTPUT_DIRECTORY}/${JSON_FILE}
	)
endforeach()