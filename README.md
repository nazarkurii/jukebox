## 🎵 Jukebox CLI

### 🚀 Running the Application

Before running the program, you must set the `JUKEBOX__TRACKS_PATH` environment variable.  
It should point to a JSON configuration file containing track definitions.

#### ▶️ Example

```bash
export JUKEBOX__TRACKS_PATH=./jukebox/config/example.json
go run cmd/main.go
```

### 📁 Configuration

The application loads tracks from a JSON file.

Example file location:

```
./jukebox/config/example.json
```

#### Example config structure

```json
{
    "tracks": [
        {
            "type": "standard",
            "artist": "Queen",
            "title": "Bohemian Rhapsody",
            "duration":5,
            "price": 1.50
        },
        {
            "type": "standard",
            "artist": "AC/DC",
             "duration":5,
            "title": "Back In Black",
            "price": 1.20
        },
        {
            "type": "vip",
            "artist": "Miles Davis",
            "title": "Kind of Blue",
             "duration":5,
            "price": 3.00,
            "vip_message": "Эксклюзивная запись. Только для ценителей."
        }
    ]
}
```

### 🎮 Usage

1. Run the application  
2. Select a track:
   - by number (`1`, `2`, `3`, ...)
   - or by name  (prefix search)
3. Insert coins (e.g. `0.25`, `1.00`)  
4. Wait for playback to complete  


### 💰 Supported Coin Denominations

```
0.01 0.05 0.10 0.25 0.50 1.00
```

Invalid coins will be rejected.

### ⌨️ Commands

- `cancel` — cancel current selection and return inserted coins  
- `history` — show previously played tracks  

