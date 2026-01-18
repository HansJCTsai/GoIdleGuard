# GoIdleGuard

> 🚫 **一個輕量級、跨平台的防止系統閒置工具，專為開發者與辦公環境設計。**

**GoIdleGuard** 是一個以後端服務 (Daemon) 形式運行的工具，常駐於系統匣 (System Tray)。它能根據您設定的「工作排程」，在特定時段內自動偵測並模擬使用者活動（如滑鼠移動），有效防止電腦進入睡眠或螢幕保護模式，同時提供直覺的狀態顯示與日誌監控功能。

---

## ✨ 主要功能 (Features)

* 📅 **智慧工作排程**
    可自定義每日工作時段（例如 09:00-18:00），非工作時間（如午休或下班後）自動暫停，讓電腦恢復正常休眠。
* 🖱️ **多種模擬模式**
    支援 **滑鼠移動 (Mouse)**、**鍵盤按鍵 (Key)** 或 **混合模式 (Mixed)** 來防止閒置。
* 🖥️ **雙平台支援**
    完美支援 **Windows** (包含 `.exe` 圖示與隱藏視窗) 與 **macOS** (App Bundle)。
* 📊 **系統匣整合**
    常駐右下角系統列，提供「即時日誌監控 (Live Logs)」、「快速設定」與「關於」介面。
* 📝 **自動日誌輪替**
    內建 Log Rotation 機制，自動切割與壓縮日誌檔案，防止佔用過多硬碟空間。
* 🛡️ **穩定性設計**
    包含錯誤捕捉與自動恢復機制，確保長時間穩定運行。

---

## 🛠️ 安裝與建置 (Build & Installation)

本專案使用 `Makefile` 進行自動化建置。

### 前置需求
* [Go 1.20+](https://go.dev/dl/)
* **Make** 工具 (Windows 使用者可安裝 MinGW 或直接使用 Git Bash)
* **(Windows 編譯專用)** `rsrc` 工具：用於嵌入執行檔圖示。
    ```bash
    go install [github.com/akavel/rsrc@latest](https://github.com/akavel/rsrc@latest)
    ```

### 建置指令
下載專案後，在終端機執行以下指令：

#### 🪟 Windows
產生帶有圖示的背景執行檔 (`bin/app-daemon.exe`)：
```bash
make windows-daemon
```

#### 🍎 macOS
打包成標準應用程式 (`bin/GoIdleGuard.app`)，可直接放入應用程式資料夾：
```bash
make macos-app
```

---

## ⚙️ 設定說明 (Configuration)

程式啟動時會自動讀取目錄下的 `config.yaml` 設定檔。您也可以直接在系統匣選單中點選 **"Settings"** 進行修改。

以下是設定檔的完整範例與說明：

```yaml
# config.yaml 範例
version:
  name: PreventIdleApp
  version: "1.0.0"

# 閒置偵測設定
idlePrevention:
  enabled: true       # 總開關
  interval: "5s"      # 閒置判定時間：當系統閒置超過此時間，觸發防閒置動作
  mode: "mouse"       # 運作模式：mouse (滑鼠微動), key (模擬按鍵), mixed (混合)

# 日誌設定
logging:
  level: "info"       # 日誌等級：debug, info, warn, error

# 工作排程 (Work Schedule)
# 注意：只有在定義的時段內，程式才會運作。
# 規則：
# 1. 空陣列 [] 代表當天完全不運作。
# 2. 若無該星期的設定區塊，則預設為「全天運作」。
workSchedule:
  monday:
    - start: "08:00"
      end: "12:00"    # 上午班
    - start: "13:00"  # 午休結束 (12:00-13:00 電腦可正常休眠)
      end: "17:00"    # 下午班
  tuesday:
    - start: "08:00"
      end: "17:00"
  # ... 其他星期可依此類推
  sunday: []          # 週日設定為空陣列，代表不運作
```

---

## 🚀 使用方式 (Usage)

1.  **啟動程式**
    直接執行編譯好的檔案：
    * **Windows**: 執行 `bin/app-daemon.exe`
    * **macOS**: 執行 `bin/GoIdleGuard.app`

2.  **背景執行**
    程式啟動後會自動縮小至系統匣 (System Tray) 區域維持背景運作，不會干擾您的工作列。

3.  **功能選單**
    在系統匣圖示上點擊（右鍵或左鍵）即可開啟選單：
    * **Show Logs (Live)**：開啟即時日誌視窗，查看程式目前的運作狀態與模擬紀錄。
    * **Settings**：直接開啟 `config.yaml` 設定檔進行編輯。
    * **Quit**：完全終止並關閉程式。

---

## 📂 專案結構 (Project Structure)

```text
GoIdleGuard/
├── bin/                 # 編譯輸出目錄 (Binary output)
├── cmd/
│   ├── daemon/          # 背景服務主程式 (Main Logic)
│   │   ├── main.go      # 程式進入點
│   │   ├── icon.ico     # Windows 圖示資源
│   │   └── icon.icns    # macOS 圖示資源
│   └── gui/             # (Optional) 設定介面程式
├── internal/
│   ├── config/          # 設定檔讀取與解析
│   ├── preventidle/     # 防閒置核心邏輯 (Mouse/Key Simulation)
│   └── schedule/        # 工作排程計算器
├── pkg/
│   └── logger/          # 日誌封裝 (Lumberjack 整合)
├── config.yaml          # 使用者設定檔
├── Makefile             # 自動化建置腳本
└── README.md            # 專案說明文件
```

---

## 🤝 貢獻 (Contributing)

我們非常歡迎您參與專案的改善！無論是回報 Bug、提出新功能建議，或是直接提交程式碼。

**參與流程：**

1. **Fork** 本專案到您的 GitHub 帳號。
2. **Clone** 專案到您的電腦：
    ```bash
    git clone [https://github.com/HansJCTsai/GoIdleGuard.git](https://github.com/HansJCTsai/GoIdleGuard.git)
    ```
3.  建立您的功能分支 (**Feature Branch**)：
    ```bash
    git checkout -b feature/AmazingFeature
    ```
4.  提交您的修改 (**Commit**)：
    ```bash
    git commit -m 'Add some AmazingFeature'
    ```
5.  推送到您的分支 (**Push**)：
    ```bash
    git push origin feature/AmazingFeature
    ```
6.  回到 GitHub 開啟 **Pull Request**，並簡述您的修改內容。