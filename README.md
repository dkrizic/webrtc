# WebRTC

This project is a test for using a purely Web based solution to use SIP telephony.

```mermaid
graph TD
    style P fill:lightgreen
    style AG fill:lightblue
    style ADM shape:rectangle, fill:lightblue
    style CUS shape:rectangle, fill:lightblue
    style CA fill:lightcoral
    
    AG[👤 Agent]:::person
    ADM[👤 Admin]
    CUS[👤 Customer]

    subgraph P[Portal]
      subgraph F[Frontend]
        UI[UI/Forms/HTMX]
        Telephony[Telephony WebRTC]
      end
      subgraph B[Backend]
        BL[Business Logic]
        SIG[Signaling Service<br/>WebSocket]
        SIP[WebRTC to SIP Gateway]
      end
    end

    subgraph DB["🗄️Database"]
      Agent
      Call
      Customer
      Form
    end
    
    subgraph CA[Clarity Adapter]
      CAAPI[Generic API]
      CLSDK[Clarity SDK]
    end

    subgraph PBX[Clarity]
      CSIP[SIP Server]
      CAPI[Clarity API]
    end

    AG-->|Browser|F
    ADM-->|Browser|F
    CUS-->|Browser|F
    CUS-.->|API Access|BL
    BL-->|SQL|DB
    UI-->|GraphQL|BL
    Telephony-->|WebSocket|SIG
    SIG-->|SDP Offer/Answer|SIP
    Telephony-->|WebRTC ICE|SIP
    SIP-->|SIP|CSIP
    BL-->|API|CAAPI
    CAAPI-->|SDK|CLSDK
    CLSDK-->|API|CAPI
```

## Backend

The backend is written in Go and acts as a WebSocket server to handle signaling for WebRTC and SIP. It also provides a REST API for the frontend to interact with.

## Frontend

The frontend is a simple HTML/JavaScript application that allows users to make and receive calls using WebRTC. It connects to the backend via WebSocket for signaling and uses the REST API for other interactions.
