// dashboard/app/dashboard/[guildId]/clock/page.tsx
"use client";

import { useEffect, useRef, useState } from "react";
import { useParams } from "next/navigation";
import { Plus, Trash2, Pencil, X } from "lucide-react";
import styles from "@/app/dashboard/[guildId]/dashboard.module.css";
import {
  ShrimpyAPI,
  ClockChannel,
  TimezoneOption,
  DiscordChannel,
  CLOCK_FORMATS,
} from "@/lib/api";
import Dropdown from "@/components/Dropdown";
import { useToast } from "@/hooks/useToast";
import { SkeletonCard, SkeletonHeader } from "@/components/Skeleton/Skeleton";

const DEFAULT_FORMAT = CLOCK_FORMATS[0].value;
const DEFAULT_TIMEZONE = "UTC";

const SOURCE_OPTIONS = [
  { value: "create", label: "Create a channel for me" },
  { value: "existing", label: "Use an existing voice channel" },
];

const FORMAT_OPTIONS = CLOCK_FORMATS.map(f => ({
  value: f.value,
  label: f.label,
  group: f.kind === "date" ? "Date" : "Time",
}));

const formatLabel = (format: string) =>
  CLOCK_FORMATS.find(f => f.value === format)?.label ?? format;

export default function ClockPage() {
  const params = useParams();
  const guildId = params?.guildId as string;
  const { showToast, updateToast } = useToast();

  const [clocks, setClocks] = useState<ClockChannel[]>([]);
  const [timezones, setTimezones] = useState<TimezoneOption[]>([]);
  const [voiceChannels, setVoiceChannels] = useState<DiscordChannel[]>([]);
  const [categories, setCategories] = useState<DiscordChannel[]>([]);
  const [loading, setLoading] = useState(true);

  // Form state. editingId is set while editing an existing clock.
  const [editingId, setEditingId] = useState<string | null>(null);
  const [format, setFormat] = useState(DEFAULT_FORMAT);
  const [timezone, setTimezone] = useState(DEFAULT_TIMEZONE);
  const [emoji, setEmoji] = useState("");
  const [source, setSource] = useState("create");
  const [voiceChannelId, setVoiceChannelId] = useState("");
  const [categoryId, setCategoryId] = useState("");

  const isSubmitting = useRef(false);
  const guildIdRef = useRef(guildId);
  useEffect(() => {
    guildIdRef.current = guildId;
  }, [guildId]);

  useEffect(() => {
    async function loadData() {
      setLoading(true);
      try {
        const [clockData, tzData, voiceData, categoryData] = await Promise.all([
          ShrimpyAPI.getClockChannels(guildId),
          ShrimpyAPI.getClockTimezones(guildId),
          ShrimpyAPI.getDiscordVoiceChannels(guildId),
          ShrimpyAPI.getDiscordChannelGroups(guildId),
        ]);
        setClocks(clockData);
        setTimezones(tzData);
        setVoiceChannels(voiceData);
        setCategories(categoryData);
      } catch (err) {
        console.error(err);
        showToast("Failed to load clock channels.", "error");
      } finally {
        setLoading(false);
      }
    }
    loadData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [guildId]);

  const resetForm = () => {
    setEditingId(null);
    setFormat(DEFAULT_FORMAT);
    setTimezone(DEFAULT_TIMEZONE);
    setEmoji("");
    setSource("create");
    setVoiceChannelId("");
    setCategoryId("");
  };

  const handleEdit = (clock: ClockChannel) => {
    setEditingId(clock.id);
    setFormat(clock.format);
    setTimezone(clock.timezone);
    setEmoji(clock.emoji ?? "");
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (isSubmitting.current) return;
    if (!editingId && source === "existing" && !voiceChannelId) {
      showToast("Pick a voice channel to link.", "error");
      return;
    }
    isSubmitting.current = true;

    // Snapshot, then reset the form right away so the next clock can be entered
    // without waiting on the network.
    const snapshot = {
      editingId,
      format,
      timezone,
      emoji: emoji.trim(),
      source,
      voiceChannelId,
      categoryId,
    };
    const requestGuildId = guildId;
    resetForm();
    isSubmitting.current = false;

    const toastId = showToast(snapshot.editingId ? "Updating clock…" : "Setting up clock…", "loading");
    try {
      if (snapshot.editingId) {
        const saved = await ShrimpyAPI.updateClockChannel(requestGuildId, snapshot.editingId, {
          format: snapshot.format,
          timezone: snapshot.timezone,
          emoji: snapshot.emoji,
        });
        if (guildIdRef.current === requestGuildId) {
          setClocks(prev => prev.map(c => (c.id === saved.id ? saved : c)));
        }
        updateToast(toastId, "Clock updated.", "success");
      } else {
        const created = await ShrimpyAPI.createClockChannel(requestGuildId, {
          format: snapshot.format,
          timezone: snapshot.timezone,
          emoji: snapshot.emoji,
          channelId: snapshot.source === "existing" ? snapshot.voiceChannelId : null,
          categoryId: snapshot.source === "create" ? snapshot.categoryId || null : null,
        });
        if (guildIdRef.current === requestGuildId) {
          setClocks(prev => [...prev.filter(c => c.id !== created.id), created]);
        }
        updateToast(toastId, "Clock added.", "success");
      }
    } catch (err) {
      console.error(err);
      updateToast(toastId, snapshot.editingId ? "Failed to update clock." : "Failed to add clock.", "error");
    }
  };

  const handleDelete = async (clock: ClockChannel) => {
    // A bot-created channel exists only for this clock, so offer to remove it too.
    const deleteChannel = clock.managed
      ? window.confirm("Also delete the Discord channel the bot created for this clock?\n\nOK = delete the channel, Cancel = keep it.")
      : false;
    try {
      await ShrimpyAPI.deleteClockChannel(guildId, clock.id, deleteChannel);
      setClocks(prev => prev.filter(c => c.id !== clock.id));
      if (editingId === clock.id) resetForm();
      showToast("Clock removed.", "success");
    } catch (err) {
      console.error(err);
      showToast("Failed to remove clock.", "error");
    }
  };

  if (loading) return (
    <div>
      <SkeletonHeader />
      <div style={{ display: "flex", flexDirection: "column", gap: "var(--space-6)", maxWidth: 720 }}>
        <SkeletonCard fields={3} />
        <SkeletonCard fields={4} />
      </div>
    </div>
  );

  const tzLabel = (id: string) => timezones.find(t => t.id === id)?.label ?? id;
  const tzOptions = timezones.map(t => ({ value: t.id, label: t.label }));
  const linkedIds = new Set(clocks.map(c => c.channelId));
  const voiceOptions = voiceChannels
    .filter(c => !linkedIds.has(c.id))
    .map(c => ({ value: c.id, label: c.name, icon: "🔊" }));

  return (
    <div>
      <div className={styles.sectionHeader}>
        <h2 className={styles.sectionTitle}>Clock Channels</h2>
        <p className={styles.sectionDesc}>
          Show the current date or time in a channel name. The bot renames the channel automatically.
        </p>
      </div>

      <div style={{ display: "flex", flexDirection: "column", gap: "var(--space-6)", maxWidth: 720 }}>

        {/* Existing clocks */}
        <div className={styles.card}>
          <h3 className={styles.cardTitle}>Your Clocks</h3>
          <div style={{ display: "flex", flexDirection: "column", gap: "8px" }}>
            {clocks.length === 0 ? (
              <div style={{ color: "var(--color-text-muted)", fontSize: "12px" }}>No clocks yet. Add one below.</div>
            ) : (
              clocks.map(clock => (
                <div key={clock.id} className={styles.actionBtn} style={{ justifyContent: "space-between", cursor: "default" }}>
                  <span style={{ display: "flex", flexDirection: "column", gap: 2 }}>
                    <span style={{ fontWeight: "bold" }}>{clock.name || "Waiting for first update…"}</span>
                    <span style={{ fontSize: 11, color: "var(--color-text-muted)" }}>
                      {formatLabel(clock.format)} · {tzLabel(clock.timezone)} · {clock.managed ? "Created by bot" : "Linked channel"}
                    </span>
                  </span>
                  <span style={{ display: "flex", alignItems: "center", gap: 8 }}>
                    <button type="button" onClick={() => handleEdit(clock)} aria-label="Edit clock"
                      style={{ background: "none", border: "none", color: "var(--color-text-muted)", cursor: "pointer" }}>
                      <Pencil size={12} />
                    </button>
                    <button type="button" onClick={() => handleDelete(clock)} aria-label="Remove clock"
                      style={{ background: "none", border: "none", color: "var(--color-danger)", cursor: "pointer" }}>
                      <Trash2 size={12} />
                    </button>
                  </span>
                </div>
              ))
            )}
          </div>
        </div>

        {/* Add / edit form */}
        <form onSubmit={handleSubmit} className={styles.card}>
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
            <h3 className={styles.cardTitle}>{editingId ? "Edit Clock" : "Add a Clock"}</h3>
            {editingId && (
              <button type="button" onClick={resetForm} className={styles.actionBtn}
                style={{ padding: "6px 10px", display: "flex", alignItems: "center", gap: 4 }}>
                <X size={12} /><span>Cancel</span>
              </button>
            )}
          </div>

          <div style={{ display: "flex", flexDirection: "column", gap: "var(--space-4)" }}>
            <div className={styles.formGroup}>
              <label className={styles.label}>Display</label>
              <Dropdown value={format} onChange={setFormat} options={FORMAT_OPTIONS} />
            </div>

            <div className={styles.formGroup}>
              <label className={styles.label}>Timezone</label>
              <Dropdown value={timezone} onChange={setTimezone} options={tzOptions} placeholder="UTC" />
              <div style={{ fontSize: "11px", color: "var(--color-text-muted)", marginTop: 4 }}>
                Sorted by GMT offset. Defaults to UTC.
              </div>
            </div>

            <div className={styles.formGroup}>
              <label className={styles.label}>Emoji (optional)</label>
              <input
                className={styles.input}
                type="text"
                value={emoji}
                onChange={ev => setEmoji(ev.target.value)}
                maxLength={8}
                placeholder="Leave empty for the default 📅 / 🕐"
              />
              <div style={{ fontSize: "11px", color: "var(--color-text-muted)", marginTop: 4 }}>
                Standard emoji only; custom server emoji can&apos;t appear in channel names.
              </div>
            </div>

            {!editingId && (
              <>
                <div className={styles.formGroup}>
                  <label className={styles.label}>Channel</label>
                  <Dropdown value={source} onChange={setSource} options={SOURCE_OPTIONS} />
                </div>

                {source === "create" ? (
                  <div className={styles.formGroup}>
                    <label className={styles.label}>Category (optional)</label>
                    <Dropdown
                      value={categoryId}
                      onChange={setCategoryId}
                      options={[{ value: "", label: "No category" }, ...categories.map(c => ({ value: c.id, label: c.name }))]}
                    />
                    <div style={{ fontSize: "11px", color: "var(--color-text-muted)", marginTop: 4 }}>
                      The bot creates a voice channel members can see but not join.
                    </div>
                  </div>
                ) : (
                  <div className={styles.formGroup}>
                    <label className={styles.label}>Voice Channel</label>
                    <Dropdown
                      value={voiceChannelId}
                      onChange={setVoiceChannelId}
                      placeholder={voiceOptions.length === 0 ? "No unlinked voice channels" : "Select a voice channel..."}
                      options={voiceOptions}
                    />
                    <div style={{ fontSize: "11px", color: "var(--color-text-muted)", marginTop: 4 }}>
                      The channel&apos;s name will be replaced by the clock.
                    </div>
                  </div>
                )}
              </>
            )}
          </div>

          <button type="submit" className={styles.submitBtn} style={{ marginTop: "var(--space-4)" }}>
            {editingId ? <Pencil size={16} /> : <Plus size={16} />}
            <span>{editingId ? "Save Changes" : "Add Clock"}</span>
          </button>
        </form>

        <div style={{ fontSize: "12px", color: "var(--color-text-muted)" }}>
          Discord limits channel renames, so clocks refresh about every 5 minutes and the time may be a few
          minutes behind. The bot needs the <strong>Manage Channels</strong> permission.
        </div>
      </div>
    </div>
  );
}
