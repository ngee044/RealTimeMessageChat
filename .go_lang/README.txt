graph TD;
    subgraph Client
        A[Client 요청] -->|HTTP 요청| B[RestAPI]
    end

    subgraph RestAPI
        B -->|Publish Message| C[RabbitMQ]
    end

    subgraph MessageQueue
        C -->|Consume Message| E[Consumer]
        E -->|Store Data| D
    end

    subgraph Storage
        D[(Redis)]
    end

    subgraph Message Broker
        C[(RabbitMQ)]
    end

    subgraph Client Polling
        F[Client Polling 요청] -->|HTTP 요청| B
        B -->|Get Data| D 
    end