cmake_minimum_required(VERSION 3.18)

set(LIBRARY_NAME CommonModule)
set(CMAKE_CXX_STANDARD 23)
set(CMAKE_CXX_STANDARD_REQUIRED TRUE)

set(SOURCE_FILES
	ClientCombinedMessageExecute.cpp
	ClientCombinedMessageParsing.cpp
	ClientMessageExecute.cpp
	ClientMessageParsing.cpp
	DatabaseConnector.cpp
	DBWorker.cpp
	ServerCombinedMessageExecute.cpp
    ServerCombinedMessageParsing.cpp
    ServerMessageExecute.cpp
    ServerMessageParsing.cpp
)

set (HEADER_FILES
	ClientCombinedMessageExecute.h
	ClientCombinedMessageParsing.h
	ClientMessageExecute.h
	ClientMessageParsing.h
	DatabaseConnector.h
	DBWorker.h
	ServerCombinedMessageExecute.h
	ServerCombinedMessageParsing.h
	ServerMessageExecute.h
	ServerMessageParsing.h
	ModuleHeader.hpp
	ServerHeader.h
	ClientHeader.h
)

project(${LIBRARY_NAME} VERSION 1.0.0.0)

add_library(${LIBRARY_NAME} ${HEADER_FILES} ${SOURCE_FILES})

find_package(Boost REQUIRED COMPONENTS json)

target_include_directories(${LIBRARY_NAME} PUBLIC ${CMAKE_CURRENT_SOURCE_DIR})
target_link_libraries(${LIBRARY_NAME} PUBLIC Boost::json Utilities Thread Network Database)
