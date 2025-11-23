"use client";

import { useEffect, useState } from "react";
import { Activity, CheckCircle2, Clock, Globe, Server, X } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import { generateMockData, type MonitorData } from "@/lib/mock-data";
import { Sparkline } from "@/components/sparkline";

export function Dashboard() {
  const [data, setData] = useState<MonitorData | null>(null);
  const [loading, setLoading] = useState(true);
  const [open, setOpen] = useState(false);
  const [searchFilter, setSearchFilter] = useState<string | null>(null);

  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen((open) => !open);
      }
    };
    document.addEventListener("keydown", down);
    return () => document.removeEventListener("keydown", down);
  }, []);

  useEffect(() => {
    const load = async () => {
      setLoading(true);
      try {
        const base = process.env.NEXT_PUBLIC_BASE_PATH || "";

        let statusData = null;
        let historyData = {};

        try {
          const statusRes = await fetch(`${base}/status.json`);
          if (statusRes.ok) {
            statusData = await statusRes.json();
          }
        } catch (e) {
          console.error("Failed to fetch status", e);
        }

        try {
          const historyRes = await fetch(`${base}/history.json`);
          if (historyRes.ok) {
            historyData = await historyRes.json();
          }
        } catch (e) {
          console.error("Failed to fetch history", e);
        }

        if (statusData) {
          setData({
            ...statusData,
            history: historyData || {},
          });
        } else {
          // Fallback or error handling
          console.error("Failed to fetch status");
        }
      } catch (e) {
        console.error("Error loading data", e);
      } finally {
        setLoading(false);
      }
    };

    load();
    const interval = setInterval(load, 30000);
    return () => clearInterval(interval);
  }, []);

  const onlineCount = data?.results.filter((r) => r.up).length || 0;
  const totalCount = data?.results.length || 0;
  const uptimePercentage =
    totalCount > 0 ? ((onlineCount / totalCount) * 100).toFixed(1) : "0.0";

  return (
    <div className="flex min-h-screen flex-col bg-background text-foreground">
      {/* Main Content */}
      <main className="flex-1 space-y-4 p-4 md:p-8 pt-6">
        <div className="flex flex-col space-y-2">
          <div className="flex items-center justify-between">
            <h2 className="text-3xl font-bold tracking-tight text-blue-600">
              itsGOtime
            </h2>
            <div className="hidden text-sm text-muted-foreground md:block">
              Press{" "}
              <kbd className="pointer-events-none inline-flex h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium text-muted-foreground opacity-100">
                <span className="text-xs">⌘</span>K
              </kbd>{" "}
              to search
            </div>
          </div>
          <div className="flex items-center space-x-2 text-muted-foreground">
            <Clock className="h-4 w-4" />
            <span className="text-sm">
              Last updated:{" "}
              {loading
                ? "Updating..."
                : data?.generated_at
                ? new Date(data.generated_at).toLocaleString()
                : ""}
            </span>
          </div>
        </div>

        {/* Summary Cards */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                Total Monitors
              </CardTitle>
              <Server className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{totalCount}</div>
              <p className="text-xs text-muted-foreground">
                Active services monitoring
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Operational</CardTitle>
              <CheckCircle2 className="h-4 w-4 text-green-500" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{onlineCount}</div>
              <p className="text-xs text-muted-foreground">
                Services online right now
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                Current Uptime
              </CardTitle>
              <Activity className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{uptimePercentage}%</div>
              <p className="text-xs text-muted-foreground">
                Average availability across services
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                Avg Response
              </CardTitle>
              <Globe className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">124ms</div>
              <p className="text-xs text-muted-foreground">
                Global latency average
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Tabs and Main List */}
        <Tabs defaultValue="all" className="space-y-4">
          <div className="flex items-center justify-between">
            <TabsList>
              <TabsTrigger value="all">All Monitors</TabsTrigger>
              <TabsTrigger value="online">Online</TabsTrigger>
              <TabsTrigger value="offline">Offline</TabsTrigger>
            </TabsList>
            {searchFilter && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setSearchFilter(null)}
                className="h-8 px-2 lg:px-3"
              >
                Filtered by:{" "}
                <span className="font-bold ml-1">{searchFilter}</span>
                <X className="ml-2 h-4 w-4" />
              </Button>
            )}
          </div>
          <TabsContent value="all" className="space-y-4">
            <MonitorList
              results={data?.results || []}
              statusFilter="all"
              history={data?.history || {}}
              searchFilter={searchFilter}
            />
          </TabsContent>
          <TabsContent value="online" className="space-y-4">
            <MonitorList
              results={data?.results || []}
              statusFilter="online"
              history={data?.history || {}}
              searchFilter={searchFilter}
            />
          </TabsContent>
          <TabsContent value="offline" className="space-y-4">
            <MonitorList
              results={data?.results || []}
              statusFilter="offline"
              history={data?.history || {}}
              searchFilter={searchFilter}
            />
          </TabsContent>
        </Tabs>
      </main>

      {/* CommandDialog for the Ctrl+K search functionality */}
      <CommandDialog open={open} onOpenChange={setOpen}>
        <CommandInput placeholder="Search monitors by name or URL..." />
        <CommandList>
          <CommandEmpty>No results found.</CommandEmpty>
          <CommandGroup heading="Monitors">
            {data?.results.map((monitor) => (
              <CommandItem
                key={monitor.name}
                onSelect={() => {
                  setSearchFilter(monitor.name);
                  setOpen(false);
                }}
              >
                <div
                  className={`mr-2 h-2 w-2 rounded-full ${
                    monitor.up ? "bg-emerald-500" : "bg-red-500"
                  }`}
                />
                <span>{monitor.name}</span>
                <span className="ml-2 text-xs text-muted-foreground font-mono">
                  {monitor.url}
                </span>
              </CommandItem>
            ))}
          </CommandGroup>
        </CommandList>
      </CommandDialog>
    </div>
  );
}

function MonitorList({
  results,
  statusFilter,
  history,
  searchFilter,
}: {
  results: MonitorData["results"];
  statusFilter: "all" | "online" | "offline";
  history: Record<
    string,
    { up: boolean; timestamp: string; responseTime?: number }[]
  >;
  searchFilter: string | null;
}) {
  const filtered = results.filter((r) => {
    if (
      searchFilter &&
      !r.name.toLowerCase().includes(searchFilter.toLowerCase()) &&
      !r.url.toLowerCase().includes(searchFilter.toLowerCase())
    ) {
      return false;
    }
    if (statusFilter === "online") return r.up;
    if (statusFilter === "offline") return !r.up;
    return true;
  });

  if (filtered.length === 0) {
    return (
      <div className="flex h-[400px] flex-col items-center justify-center rounded-lg border bg-card text-card-foreground shadow-sm border-dashed">
        <div className="text-center">
          <Server className="mx-auto h-10 w-10 text-muted-foreground" />
          <h3 className="mt-4 text-lg font-semibold">No monitors found</h3>
          <p className="text-sm text-muted-foreground">
            {searchFilter
              ? `No monitors matching "${searchFilter}"`
              : "There are no monitors matching this filter."}
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="grid gap-4 md:grid-cols-1">
      <Card>
        <CardHeader className="px-6 py-4 border-b">
          <div className="grid grid-cols-12 gap-4 text-sm font-medium text-muted-foreground uppercase tracking-wider">
            <div className="col-span-4 md:col-span-3">Service</div>
            <div className="col-span-3 md:col-span-4 hidden md:block">
              Endpoint
            </div>
            <div className="col-span-4 md:col-span-2 text-right md:text-left">
              Status
            </div>
            <div className="col-span-4 md:col-span-3 text-right">Last 24h</div>
          </div>
        </CardHeader>
        <CardContent className="p-0">
          {filtered.map((monitor, i) => (
            <div
              key={monitor.name}
              className={`grid grid-cols-12 gap-4 px-6 py-4 items-center border-b last:border-0 hover:bg-muted/50 transition-colors ${
                i % 2 === 0 ? "bg-background" : "bg-background/50"
              }`}
            >
              <div className="col-span-4 md:col-span-3 font-medium flex items-center gap-2 overflow-hidden">
                <div
                  className={`h-2.5 w-2.5 rounded-full shrink-0 ${
                    monitor.up
                      ? "bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.4)]"
                      : "bg-red-500 shadow-[0_0_8px_rgba(239,68,68,0.4)]"
                  }`}
                />
                <span className="truncate" title={monitor.name}>
                  {monitor.name}
                </span>
              </div>

              <div className="col-span-3 md:col-span-4 hidden md:block text-sm text-muted-foreground truncate font-mono">
                <a
                  href={monitor.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="hover:text-primary transition-colors hover:underline"
                >
                  {monitor.url}
                </a>
              </div>

              <div className="col-span-4 md:col-span-2 flex justify-end md:justify-start">
                <Badge
                  variant={monitor.up ? "outline" : "destructive"}
                  className={
                    monitor.up
                      ? "border-emerald-500/30 text-emerald-500 bg-emerald-500/10 hover:bg-emerald-500/20"
                      : ""
                  }
                >
                  {monitor.up ? "Online" : "Offline"}
                  {!monitor.up && monitor.status !== 0
                    ? ` (${monitor.status})`
                    : ""}
                </Badge>
              </div>

              <div className="col-span-4 md:col-span-3 h-[40px] flex items-center justify-end">
                <div className="w-full max-w-[160px] h-full">
                  <Sparkline
                    data={history[monitor.name] || []}
                    color={monitor.up ? "#10b981" : "#ef4444"}
                  />
                </div>
              </div>
            </div>
          ))}
        </CardContent>
      </Card>
    </div>
  );
}
