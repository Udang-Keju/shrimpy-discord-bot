// dashboard/app/dashboard/[guildId]/clock/page.tsx
"use client";

import { useEffect, useRef, useState } from "react";
import { useParams } from "next/navigation";
import { Trash2, Pencil, ArrowLeft, ArrowRight, Check } from "lucide-react";
import styles from "@/app/dashboard/[guildId]/dashboard.module.css";
import {
  ShrimpyAPI,
  ClockChannel,
  TimezoneOption,
  DiscordChannel,
  CLOCK_FORMATS,
  CLOCK_REFRESH_OPTIONS,
} from "@/lib/api";
import Dropdown from "@/components/Dropdown";
import { useToast } from "@/hooks/useToast";
import { SkeletonCard, SkeletonHeader } from "@/components/Skeleton/Skeleton";

type Kind = "date" | "time";
const KINDS: Kind[] = ["date", "time"];
const KIND_LABEL: Record<Kind, string> = { date: "Date channel", time: "Time channel" };
const DEFAULT_EMOJI: Record<Kind, string> = { date: "📅", time: "🕐" };
const DEFAULT_TIMEZONE = "UTC";
const DEFAULT_REFRESH = 5;
const STEPS = ["Channels", "Format & timezone", "Review & save"];

const SOURCE_OPTIONS = [
  { value: "create", label: "Create a channel for me" },
  { value: "existing", label: "Use an existing voice channel" },
];

const formatsFor = (kind: Kind) => CLOCK_FORMATS.filter(f => f.kind === kind);
const formatLabel = (format: string) =>
  CLOCK_FORMATS.find(f => f.value === format)?.label ?? format;
const refreshLabel = (minutes: number) =>
  CLOCK_REFRESH_OPTIONS.find(o => o.value === minutes)?.label ?? `Every ${minutes} minutes`;

// Per-channel wizard state. `existing` is set when the server already has a clock of
// this kind: its Discord channel is fixed and only the display settings are edited.
interface ChannelDraft {
  existing: ClockChannel | null;
  source: string;
  voiceChannelId: string;
  categoryId: string;
  format: string;
  emoji: string;
}

const emptyDraft = (kind: Kind, existing: ClockChannel | null): ChannelDraft => ({
  existing,
  source: "create",
  voiceChannelId: "",
  categoryId: "",
  format: existing?.format ?? formatsFor(kind)[0].value,
  emoji: existing?.emoji ?? "",
});

export default function ClockPage() {
  const params = useParams();
  const guildId = params?.guildId as string;
  const { showToast, updateToast } = useToast();

  const [clocks, setClocks] = useState<ClockChannel[]>([]);
  const [timezones, setTimezones] = useState<TimezoneOption[]>([]);
  const [voiceChannels, setVoiceChannels] = useState<DiscordChannel[]>([]);
  const [categories, setCategories] = useState<DiscordChannel[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  // The wizard shows while editing, or when the server has no clocks yet.
  const [editing, setEditing] = useState(false);
  const [step, setStep] = useState(0);
  const [drafts, setDrafts] = useState<Record<Kind, ChannelDraft>>({
    date: emptyDraft("date", null),
    time: emptyDraft("time", null),
  });
  const [timezone, setTimezone] = useState(DEFAULT_TIMEZONE);
  const [refreshMinutes, setRefreshMinutes] = useState(DEFAULT_REFRESH);

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

  // The setup is the first date clock plus the first time clock; any others
  // (from before the wizard existed) are listed separately and can only be removed.
  const primary: Record<Kind, ClockChannel | null> = {
    date: clocks.find(c => c.kind === "date") ?? null,
    time: clocks.find(c => c.kind === "time") ?? null,
  };
  const extras = clocks.filter(c => c !== primary.date && c !== primary.time);
  const hasSetup = clocks.length > 0;
  const showWizard = editing || !hasSetup;

  const startWizard = () => {
    const anchor = primary.date ?? primary.time;
    setDrafts({
      date: emptyDraft("date", primary.date),
      time: emptyDraft("time", primary.time),
    });
    setTimezone(anchor?.timezone ?? DEFAULT_TIMEZONE);
    setRefreshMinutes(anchor?.refreshMinutes ?? DEFAULT_REFRESH);
    setStep(0);
    setEditing(true);
  };

  const cancelWizard = () => {
    setEditing(false);
    setStep(0);
  };

  const updateDraft = (kind: Kind, patch: Partial<ChannelDraft>) =>
    setDrafts(prev => ({ ...prev, [kind]: { ...prev[kind], ...patch } }));

  // Step 1 needs a channel picked for every clock that will link an existing one.
  const channelsError = (): string | null => {
    for (const kind of KINDS) {
      const d = drafts[kind];
      if (!d.existing && d.source === "existing" && !d.voiceChannelId) {
        return `Pick a voice channel for the ${kind} clock.`;
      }
    }
    const { date, time } = drafts;
    if (
      !date.existing && !time.existing &&
      date.source === "existing" && time.source === "existing" &&
      date.voiceChannelId === time.voiceChannelId
    ) {
      return "The date and time clocks need different channels.";
    }
    return null;
  };

  const goNext = () => {
    if (step === 0) {
      const err = channelsError();
      if (err) {
        showToast(err, "error");
        return;
      }
    }
    setStep(s => Math.min(s + 1, STEPS.length - 1));
  };

  const handleSave = async () => {
    if (isSubmitting.current) return;
    const err = channelsError();
    if (err) {
      showToast(err, "error");
      setStep(0);
      return;
    }
    isSubmitting.current = true;

    // Snapshot, then reset the wizard right away so it never shows stale input
    // while the requests are in flight.
    const snapshot = { drafts, timezone, refreshMinutes };
    const requestGuildId = guildId;
    cancelWizard();
    setSaving(true);

    const toastId = showToast("Saving clocks…", "loading");
    try {
      // Date first, then time. A failure partway leaves the first one saved, and
      // the reload below shows exactly what was applied.
      for (const kind of KINDS) {
        const d = snapshot.drafts[kind];
        const emoji = d.emoji.trim();
        if (d.existing) {
          await ShrimpyAPI.updateClockChannel(requestGuildId, d.existing.id, {
            format: d.format,
            timezone: snapshot.timezone,
            refreshMinutes: snapshot.refreshMinutes,
            emoji,
          });
        } else {
          await ShrimpyAPI.createClockChannel(requestGuildId, {
            format: d.format,
            timezone: snapshot.timezone,
            refreshMinutes: snapshot.refreshMinutes,
            emoji,
            channelId: d.source === "existing" ? d.voiceChannelId : null,
            categoryId: d.source === "create" ? d.categoryId || null : null,
          });
        }
      }
      updateToast(toastId, "Clock channels saved.", "success");
    } catch (e) {
      console.error(e);
      updateToast(toastId, "Failed to save clock channels. Check what was applied below.", "error");
    } finally {
      isSubmitting.current = false;
      try {
        const fresh = await ShrimpyAPI.getClockChannels(requestGuildId);
        if (guildIdRef.current === requestGuildId) setClocks(fresh);
      } catch (e) {
        console.error(e);
      }
      if (guildIdRef.current === requestGuildId) setSaving(false);
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
      showToast("Clock removed.", "success");
    } catch (err) {
      console.error(err);
      showToast("Failed to remove clock.", "error");
    }
  };

  const handleRemoveSetup = async () => {
    if (!window.confirm("Remove the date and time clocks? You will be asked about bot-created channels next.")) return;
    for (const clock of [primary.date, primary.time].filter((c): c is ClockChannel => c !== null)) {
      await handleDelete(clock);
    }
  };

  if (loading || saving) return (
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
  const hint = { fontSize: "11px", color: "var(--color-text-muted)", marginTop: 4 } as const;

  const channelDescription = (kind: Kind) => {
    const d = drafts[kind];
    if (d.existing) return "Keeps its current Discord channel";
    if (d.source === "existing") {
      return `Linked: ${voiceChannels.find(c => c.id === d.voiceChannelId)?.name ?? "voice channel"}`;
    }
    const category = categories.find(c => c.id === d.categoryId)?.name;
    return `Bot creates a view-only voice channel${category ? ` under ${category}` : ""}`;
  };

  return (
    <div>
      <div className={styles.sectionHeader}>
        <h2 className={styles.sectionTitle}>Clock Channels</h2>
        <p className={styles.sectionDesc}>
          Show the current date and time in two channel names. The bot renames them automatically.
        </p>
      </div>

      <div style={{ display: "flex", flexDirection: "column", gap: "var(--space-6)", maxWidth: 720 }}>

        {showWizard ? (
          <div className={styles.card}>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
              <h3 className={styles.cardTitle}>
                Step {step + 1} of {STEPS.length}: {STEPS[step]}
              </h3>
              {hasSetup && (
                <button type="button" onClick={cancelWizard} className={styles.actionBtn}
                  style={{ padding: "6px 10px" }}>
                  Cancel
                </button>
              )}
            </div>

            <div style={{ display: "flex", flexDirection: "column", gap: "var(--space-4)" }}>
              {step === 0 && KINDS.map(kind => {
                const d = drafts[kind];
                return (
                  <div key={kind} className={styles.formGroup}>
                    <label className={styles.label}>{KIND_LABEL[kind]}</label>
                    {d.existing ? (
                      <div style={hint}>
                        Already set up as &ldquo;{d.existing.name || "waiting for first update"}&rdquo;. Its channel stays the same.
                      </div>
                    ) : (
                      <>
                        <Dropdown value={d.source} onChange={v => updateDraft(kind, { source: v })} options={SOURCE_OPTIONS} />
                        {d.source === "create" ? (
                          <div style={{ marginTop: 8 }}>
                            <Dropdown
                              value={d.categoryId}
                              onChange={v => updateDraft(kind, { categoryId: v })}
                              options={[{ value: "", label: "No category" }, ...categories.map(c => ({ value: c.id, label: c.name }))]}
                            />
                            <div style={hint}>The bot creates a voice channel members can see but not join.</div>
                          </div>
                        ) : (
                          <div style={{ marginTop: 8 }}>
                            <Dropdown
                              value={d.voiceChannelId}
                              onChange={v => updateDraft(kind, { voiceChannelId: v })}
                              placeholder={voiceOptions.length === 0 ? "No unlinked voice channels" : "Select a voice channel..."}
                              options={voiceOptions}
                            />
                            <div style={hint}>The channel&apos;s name will be replaced by the clock.</div>
                          </div>
                        )}
                      </>
                    )}
                  </div>
                );
              })}

              {step === 1 && (
                <>
                  {KINDS.map(kind => {
                    const d = drafts[kind];
                    return (
                      <div key={kind} className={styles.formGroup}>
                        <label className={styles.label}>{KIND_LABEL[kind]} format</label>
                        <Dropdown
                          value={d.format}
                          onChange={v => updateDraft(kind, { format: v })}
                          options={formatsFor(kind).map(f => ({ value: f.value, label: f.label }))}
                        />
                        <input
                          className={styles.input}
                          style={{ marginTop: 8 }}
                          type="text"
                          value={d.emoji}
                          onChange={ev => updateDraft(kind, { emoji: ev.target.value })}
                          maxLength={8}
                          placeholder={`Emoji (optional, default ${DEFAULT_EMOJI[kind]})`}
                        />
                      </div>
                    );
                  })}
                  <div style={hint}>Standard emoji only; custom server emoji can&apos;t appear in channel names.</div>

                  <div className={styles.formGroup}>
                    <label className={styles.label}>Timezone</label>
                    <Dropdown value={timezone} onChange={setTimezone} options={tzOptions} placeholder="UTC" />
                    <div style={hint}>Shared by both channels. Sorted by GMT offset. Defaults to UTC.</div>
                  </div>

                  <div className={styles.formGroup}>
                    <label className={styles.label}>Refresh</label>
                    <Dropdown
                      value={String(refreshMinutes)}
                      onChange={v => setRefreshMinutes(Number(v))}
                      options={CLOCK_REFRESH_OPTIONS.map(o => ({ value: String(o.value), label: o.label }))}
                    />
                    <div style={hint}>
                      Discord allows about 2 renames per 10 minutes per channel, so 5 minutes is the fastest. The time
                      can be a few minutes behind.
                    </div>
                  </div>
                </>
              )}

              {step === 2 && (
                <>
                  {KINDS.map(kind => {
                    const d = drafts[kind];
                    return (
                      <div key={kind} className={styles.actionBtn} style={{ cursor: "default", flexDirection: "column", alignItems: "flex-start", gap: 2 }}>
                        <span style={{ fontWeight: "bold" }}>
                          {KIND_LABEL[kind]}: {d.emoji.trim() || DEFAULT_EMOJI[kind]} {formatLabel(d.format)}
                        </span>
                        <span style={{ fontSize: 11, color: "var(--color-text-muted)" }}>{channelDescription(kind)}</span>
                      </div>
                    );
                  })}
                  <div style={hint}>
                    Timezone: {tzLabel(timezone)} · {refreshLabel(refreshMinutes)}
                  </div>
                </>
              )}
            </div>

            <div style={{ display: "flex", justifyContent: "space-between", marginTop: "var(--space-4)" }}>
              <button type="button" className={styles.actionBtn} disabled={step === 0}
                onClick={() => setStep(s => Math.max(s - 1, 0))}
                style={{ padding: "8px 14px", display: "flex", alignItems: "center", gap: 6 }}>
                <ArrowLeft size={14} /><span>Back</span>
              </button>
              {step < STEPS.length - 1 ? (
                <button type="button" className={styles.submitBtn} onClick={goNext}>
                  <span>Next</span><ArrowRight size={16} />
                </button>
              ) : (
                <button type="button" className={styles.submitBtn} onClick={handleSave}>
                  <Check size={16} /><span>Save</span>
                </button>
              )}
            </div>
          </div>
        ) : (
          <div className={styles.card}>
            <h3 className={styles.cardTitle}>Your Clocks</h3>
            <div style={{ display: "flex", flexDirection: "column", gap: "8px" }}>
              {KINDS.map(kind => {
                const clock = primary[kind];
                return (
                  <div key={kind} className={styles.actionBtn} style={{ cursor: "default", flexDirection: "column", alignItems: "flex-start", gap: 2 }}>
                    <span style={{ fontWeight: "bold" }}>
                      {clock ? clock.name || "Waiting for first update…" : `No ${kind} clock`}
                    </span>
                    {clock && (
                      <span style={{ fontSize: 11, color: "var(--color-text-muted)" }}>
                        {formatLabel(clock.format)} · {tzLabel(clock.timezone)} · {refreshLabel(clock.refreshMinutes)} · {clock.managed ? "Created by bot" : "Linked channel"}
                      </span>
                    )}
                  </div>
                );
              })}
            </div>
            <div style={{ display: "flex", gap: 8, marginTop: "var(--space-4)" }}>
              <button type="button" className={styles.submitBtn} onClick={startWizard}>
                <Pencil size={16} /><span>Edit</span>
              </button>
              <button type="button" className={styles.actionBtn} onClick={handleRemoveSetup}
                style={{ padding: "8px 14px", display: "flex", alignItems: "center", gap: 6, color: "var(--color-danger)" }}>
                <Trash2 size={14} /><span>Remove</span>
              </button>
            </div>
          </div>
        )}

        {extras.length > 0 && !showWizard && (
          <div className={styles.card}>
            <h3 className={styles.cardTitle}>Other Clocks</h3>
            <div style={hint}>These were added before the guided setup and can only be removed.</div>
            <div style={{ display: "flex", flexDirection: "column", gap: "8px", marginTop: 8 }}>
              {extras.map(clock => (
                <div key={clock.id} className={styles.actionBtn} style={{ justifyContent: "space-between", cursor: "default" }}>
                  <span style={{ display: "flex", flexDirection: "column", gap: 2 }}>
                    <span style={{ fontWeight: "bold" }}>{clock.name || "Waiting for first update…"}</span>
                    <span style={{ fontSize: 11, color: "var(--color-text-muted)" }}>
                      {formatLabel(clock.format)} · {tzLabel(clock.timezone)} · {refreshLabel(clock.refreshMinutes)}
                    </span>
                  </span>
                  <button type="button" onClick={() => handleDelete(clock)} aria-label="Remove clock"
                    style={{ background: "none", border: "none", color: "var(--color-danger)", cursor: "pointer" }}>
                    <Trash2 size={12} />
                  </button>
                </div>
              ))}
            </div>
          </div>
        )}

        <div style={{ fontSize: "12px", color: "var(--color-text-muted)" }}>
          The bot needs the <strong>Manage Channels</strong> permission.
        </div>
      </div>
    </div>
  );
}
