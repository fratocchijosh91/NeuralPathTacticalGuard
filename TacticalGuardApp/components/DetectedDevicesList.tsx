import { View, Text } from "react-native";
import { Smartphone, Signal } from "lucide-react-native";
import type { DetectedDevice } from "../lib/api";

type Props = {
  devices: DetectedDevice[];
  loading?: boolean;
  error?: string | null;
};

function deviceLabel(type: string): string {
  const t = type.toUpperCase();
  if (t === "IPHONE") return "iPhone Hotspot";
  if (t === "ANDROID") return "Android Hotspot";
  return type || "Dispositivo";
}

function formatTime(iso: string): string {
  try {
    const d = new Date(iso);
    if (Number.isNaN(d.getTime())) return iso;
    return d.toLocaleString("it-IT", {
      day: "2-digit",
      month: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
    });
  } catch {
    return iso;
  }
}

export function DetectedDevicesList({ devices, loading, error }: Props) {
  const neon = "#39ff14";

  if (error) {
    return (
      <View className="mt-2 rounded-xl border border-red-900/60 bg-red-950/30 p-4">
        <Text className="font-mono text-sm text-red-300">{error}</Text>
      </View>
    );
  }

  if (loading && devices.length === 0) {
    return (
      <View className="mt-2 items-center justify-center py-8">
        <Text className="font-mono text-cyber-neon-dim">Caricamento dispositivi…</Text>
      </View>
    );
  }

  if (devices.length === 0) {
    return (
      <View className="mt-2 rounded-xl border border-dashed border-cyber-border bg-cyber-panel/50 p-6">
        <Text className="text-center font-mono text-sm text-cyber-muted">
          Nessun dispositivo in elenco. Aggiorna il file sul server (
          <Text className="text-cyber-neon-dim">NP_DETECTED_DEVICES_PATH</Text>) oppure
          verifica il deploy.
        </Text>
      </View>
    );
  }

  return (
    <View className="gap-2">
      {devices.map((item) => (
        <View
          key={item.id || `${item.type}-${item.last_seen}`}
          className="flex-row items-stretch rounded-lg border border-cyber-border bg-cyber-panel px-3 py-3"
        >
          <View className="mr-3 items-center justify-center border-r border-cyber-border pr-3">
            <Smartphone color={item.online ? neon : "#6b7280"} size={22} />
          </View>
          <View className="flex-1">
            <Text className="font-mono text-base font-semibold text-cyber-neon">
              {deviceLabel(item.type)}
            </Text>
            <View className="mt-1 flex-row items-center gap-2">
              <Signal color={item.online ? neon : "#ef4444"} size={14} />
              <Text
                className={`font-mono text-xs ${item.online ? "text-cyber-neon-dim" : "text-red-400"}`}
              >
                {item.online ? "ONLINE" : "OFFLINE"}
              </Text>
            </View>
            <Text className="mt-1 font-mono text-[10px] text-cyber-muted">
              Starlink {item.starlink_ms ?? "—"} ms · Device {item.device_ms ?? "—"} ms ·{" "}
              {formatTime(item.last_seen)}
            </Text>
          </View>
        </View>
      ))}
    </View>
  );
}
