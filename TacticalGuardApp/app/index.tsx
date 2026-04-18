import { useCallback, useEffect, useState } from "react";
import {
  View,
  Text,
  ScrollView,
  RefreshControl,
  Pressable,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { Link } from "expo-router";
import { Activity, RadioTower } from "lucide-react-native";
import { ServerStatusCard } from "@/components/ServerStatusCard";
import { DetectedDevicesList } from "@/components/DetectedDevicesList";
import {
  fetchDetectedDevices,
  fetchServerConnected,
  getRailwayBaseUrl,
  type DetectedDevice,
} from "@/lib/api";

export default function HomeScreen() {
  const baseUrl = getRailwayBaseUrl();
  const [connected, setConnected] = useState<boolean | null>(null);
  const [devices, setDevices] = useState<DetectedDevice[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [devicesError, setDevicesError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setDevicesError(null);
    const [ok, devs] = await Promise.all([
      fetchServerConnected(baseUrl),
      fetchDetectedDevices(baseUrl).catch((e: Error) => {
        setDevicesError(e.message ?? "Errore caricamento dispositivi");
        return [] as DetectedDevice[];
      }),
    ]);
    setConnected(ok);
    setDevices(devs);
  }, [baseUrl]);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setLoading(true);
      await load();
      if (!cancelled) setLoading(false);
    })();
    return () => {
      cancelled = true;
    };
  }, [load]);

  const onRefresh = useCallback(async () => {
    setRefreshing(true);
    await load();
    setRefreshing(false);
  }, [load]);

  return (
    <SafeAreaView className="flex-1 bg-cyber-bg" edges={["bottom", "left", "right"]}>
      <ScrollView
        className="flex-1 px-4 pt-2"
        refreshControl={
          <RefreshControl
            refreshing={refreshing}
            onRefresh={onRefresh}
            tintColor="#39ff14"
            colors={["#39ff14"]}
          />
        }
      >
        <View className="mb-4 flex-row items-center justify-between">
          <View className="flex-row items-center gap-2">
            <RadioTower color="#39ff14" size={26} strokeWidth={2.2} />
            <View>
              <Text className="font-mono text-lg font-bold tracking-tight text-cyber-neon">
                TACTICAL GUARD
              </Text>
              <Text className="font-mono text-[10px] uppercase text-cyber-muted">
                NeuralPath · link operativo
              </Text>
            </View>
          </View>
          <Link href="/diagnostics" asChild>
            <Pressable className="flex-row items-center gap-1 rounded border border-cyber-border bg-cyber-panel px-2 py-2">
              <Activity color="#39ff14" size={18} />
              <Text className="font-mono text-[10px] text-cyber-neon-dim">TEST</Text>
            </Pressable>
          </Link>
        </View>

        <ServerStatusCard
          connected={connected}
          loading={loading && connected === null}
          baseUrl={baseUrl}
        />

        <View className="mt-6">
          <Text className="mb-2 font-mono text-xs uppercase tracking-[0.2em] text-cyber-neon-dim">
            Ultimi dispositivi rilevati
          </Text>
          <DetectedDevicesList
            devices={devices}
            loading={loading}
            error={devicesError}
          />
        </View>

        <View className="h-10" />
      </ScrollView>
    </SafeAreaView>
  );
}
