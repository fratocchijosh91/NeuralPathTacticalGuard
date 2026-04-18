import { View, Text, ActivityIndicator } from "react-native";
import { Server, WifiOff } from "lucide-react-native";

type Props = {
  connected: boolean | null;
  loading?: boolean;
  baseUrl: string;
};

export function ServerStatusCard({ connected, loading, baseUrl }: Props) {
  const neon = "#39ff14";
  const isOn = connected === true;

  return (
    <View className="rounded-xl border border-cyber-border bg-cyber-panel p-4">
      <View className="mb-2 flex-row items-center justify-between">
        <Text className="font-mono text-xs uppercase tracking-widest text-cyber-neon-dim">
          Server Railway
        </Text>
        {loading ? (
          <ActivityIndicator color={neon} size="small" />
        ) : isOn ? (
          <Server color={neon} size={22} strokeWidth={2.2} />
        ) : (
          <WifiOff color="#ff4444" size={22} strokeWidth={2.2} />
        )}
      </View>
      <Text
        className={`font-mono text-2xl font-bold ${isOn ? "text-cyber-neon" : "text-red-400"}`}
      >
        {loading ? "…" : isOn ? "CONNESSO" : "DISCONNESSO"}
      </Text>
      <Text className="mt-2 font-mono text-[10px] text-cyber-muted" numberOfLines={2}>
        {baseUrl}
      </Text>
    </View>
  );
}
