# 🛰️ NeuralPath Tactical Guard

<p align="center">
  <!-- INSERISCI QUI IL LINK AL TUO SCREENSHOT VERO QUANDO LO CARICHI SU GITHUB -->
  <img src="https://via.placeholder.com/800x600.png?text=Inserisci+qui+lo+screenshot+della+tua+UI+Tattica" alt="NeuralPath Tactical Guard UI">
</p>

**NeuralPath Tactical Guard** is a professional-grade, tactical network monitoring tool written in Go and Fyne. Designed for mission-critical setups, it provides real-time latency tracking and failover awareness between a Primary connection (e.g. Starlink) and a Secondary cellular backup (iPhone/Android Hotspot).

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/GUI-Fyne_v2-1f415c?style=for-the-badge" alt="Fyne Framework">
  <img src="https://img.shields.io/badge/Platform-Windows_x64-blue?style=for-the-badge&logo=windows" alt="Platform">
  <a href="LICENSE.txt"><img src="https://img.shields.io/badge/License-Open_Core-brightgreen?style=for-the-badge" alt="License"></a>
</p>

## ✨ Key Features

- **Tactical UI:** A sleek, high-visibility, dark-mode interface built for quick situational awareness.
- **Dual-Link Monitoring:** Live, ms-accurate ping tracking for two distinct network paths.
- **Dynamic Live Graphing:** Visual history of latency spikes and dropouts.
- **Cross-Platform & Lightweight:** Native, incredibly fast execution powered by Go.
- **Open-Core Architecture:** The essential dashboard is open and free, while advanced data-logging features are reserved for project Sponsors.

---

## ⚖️ The Open-Core Model: Free vs PRO

Our philosophy is simple: **Core network monitoring should be accessible to everyone.** That's why the entire visual dashboard, dual-link tracking, and live graphing engine are completely free and open-source.

However, developing and maintaining a mission-critical tool takes time and resources. To sustain the project, **enterprise-grade features** (like automated reporting, historical analysis, and alert integrations) are reserved for our **GitHub Sponsors**.

### 💼 How it works
If you become a Sponsor, we will issue a secure **PRO License Key** (`NP-PRO-...`). By pasting this key into the Tactical Guard UI, the application instantly unlocks the premium tiers, expanding its capabilities from a simple monitor into a data-logging powerhouse.

| Feature Area | Free Version | PRO Version (Sponsors) |
| :--- | :---: | :---: |
| **Live Ping & Graph UI** | ✅ | ✅ |
| **Dual Network Tracking** | ✅ | ✅ |
| **Alert & Desktop Notifications** | ❌ *(Test Only)* | ✅ |
| **Session Reports (TXT/CSV)** | ❌ *(Test Only)* | ✅ |
| **Persistent Averages Analytics**| ❌ | ✅ |

💖 **[Become a GitHub Sponsor to get your PRO License Key!](#)** *(Sostituisci il # con il link alla tua pagina Sponsor)*

## 🚀 Installation 

You don't need to compile anything to start using NeuralPath Tactical Guard.

1. Head over to the [Releases](../../releases) tab.
2. Download the latest `NeuralPath-Tactical-Guard-Setup-x64.exe` (installer) or the portable ZIP.
3. Run the application and start monitoring your uplinks instantly.

## 🛠️ Build it yourself (For Developers)

If you want to compile the project from source, ensure you have [Go](https://go.dev/) installed on your machine.

```bash
git clone https://github.com/YOUR_USERNAME/NeuralPath_Lab.git
cd NeuralPath_Lab
go mod tidy

# Compilazione manuale per Windows (nascondendo il terminale)
go build -ldflags="-s -w -H windowsgui" -o NeuralPathTacticalGuard.exe .
```

*Note: For official packaging, we use the `packaging/build-release.ps1` PowerShell script which generates the app bundle and the Inno Setup installer directly.*

## 📜 License & Credits

- **Lead Developer**: Josh Fratocchi
- **Copyright** © 2026 NeuralPath Tactical
- **Core License**: Open-Source (See `LICENSE.txt`). 
- **PRO Version**: Requires a commercial/sponsor license (See `EULA.txt`).
