import { View, Text, Pressable } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { Link } from "expo-router";
import { Terminal } from "lucide-react-native";
import { getRailwayBaseUrl } from "../lib/api";

export default function DiagnosticsScreen() {
  const base = getRailwayBaseUrl();

  return (
    <SafeAreaView className="flex-1 bg-cyber-bg px-4 pt-4" edges={["bottom", "left", "right"]}>
      <View className="mb-4 flex-row items-center gap-2">
        <Terminal color="#39ff14" size={24} />
        <Text className="font-mono text-xl font-bold text-cyber-neon">Diagnostica</Text>
      </View>

      <View className="rounded-xl border border-cyber-border bg-cyber-panel p-4">
        <Text className="mb-2 font-mono text-xs uppercase text-cyber-neon-dim">Endpoint attivi</Text>
        <Text className="font-mono text-sm text-cyber-neon-dim">GET {base}/healthz</Text>
        <Text className="mt-1 font-mono text-sm text-cyber-neon-dim">
          GET {base}/v1/detected-devices
        </Text>
      </View>

      <View className="mt-4 rounded-xl border border-cyber-border bg-cyber-panel p-4">
        <Text className="mb-2 font-mono text-xs uppercase text-cyber-neon-dim">
          Variabile Expo
        </Text>
        <Text className="font-mono text-xs leading-5 text-cyber-muted">
          Imposta EXPO_PUBLIC_RAILWAY_BASE_URL per puntare a un altro deploy (senza slash finale).
        </Text>
      </View>

      <Link href="/" asChild>
        <Pressable className="mt-8 self-center rounded border border-cyber-neon bg-cyber-panel px-4 py-3 active:opacity-80">
          <Text className="font-mono text-sm font-semibold text-cyber-neon">← Torna alla home</Text>
        </Pressable>
      </Link>
    </SafeAreaView>
  );
}
