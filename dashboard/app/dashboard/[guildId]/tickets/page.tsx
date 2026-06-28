// dashboard/app/dashboard/[guildId]/tickets/page.tsx
"use client";

import { useEffect, useState, useCallback } from "react";
import { useParams } from "next/navigation";
import {
  Check,
  Lock,
  Unlock,
  Trash2,
  Download,
  Loader2,
  RefreshCw
} from "lucide-react";
import styles from "@/app/dashboard/[guildId]/dashboard.module.css";
import { ShrimpyAPI, Ticket } from "@/lib/api";
import { useToast } from "@/hooks/useToast";

export default function TicketsPage() {
  const params = useParams();
  const guildId = params?.guildId as string;
  const { showToast } = useToast();

  const [loading, setLoading] = useState(true);
  const [tickets, setTickets] = useState<Ticket[]>([]);
  const [statusFilter, setStatusFilter] = useState<'all' | 'open' | 'claimed' | 'closed'>('all');

  const loadTickets = useCallback(async () => {
    setLoading(true);
    try {
      const data = await ShrimpyAPI.listTickets(guildId);
      setTickets(data);
    } catch (err) {
      console.error("Failed to load tickets", err);
    } finally {
      setLoading(false);
    }
  }, [guildId]);

  useEffect(() => {
    const timer = setTimeout(() => {
      loadTickets();
    }, 0);
    return () => clearTimeout(timer);
  }, [loadTickets]);

  const handleClaim = async (ticketId: string) => {
    try {
      await ShrimpyAPI.claimTicket(guildId, ticketId, "StaffModerator");
      await loadTickets();
    } catch (e) {
      console.error(e);
    }
  };

  const handleClose = async (ticketId: string) => {
    try {
      await ShrimpyAPI.closeTicket(guildId, ticketId);
      await loadTickets();
    } catch (e) {
      console.error(e);
    }
  };

  const handleReopen = async (ticketId: string) => {
    try {
      await ShrimpyAPI.reopenTicket(guildId, ticketId);
      await loadTickets();
    } catch (e) {
      console.error(e);
    }
  };

  const handleArchive = async (ticketId: string) => {
    try {
      await ShrimpyAPI.archiveTicket(guildId, ticketId);
      await loadTickets();
    } catch (e) {
      console.error(e);
    }
  };

  const handleDownloadTranscript = (ticketId: string) => {
    showToast(`Downloading transcript for ${ticketId}. In production, this targets GET /api/v1/guilds/:guildId/tickets/:ticketId/transcript`, "info");
  };

  const filteredTickets = tickets.filter(t => {
    if (statusFilter === 'all') return t.status !== 'archived';
    return t.status === statusFilter;
  });

  return (
    <div>
      <div className={styles.sectionHeader}>
        <h2 className={styles.sectionTitle}>Tickets Manager</h2>
        <p className={styles.sectionDesc}>Review active customer support threads, assign claim parameters, and save transcripts.</p>
      </div>

      <div className={styles.card} style={{ gap: 'var(--space-6)' }}>
        {/* Controls Row */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: 'var(--space-4)' }}>
          <div style={{ display: 'flex', gap: '8px' }}>
            {(['all', 'open', 'claimed', 'closed'] as const).map(f => (
              <button
                key={f}
                onClick={() => setStatusFilter(f)}
                className={`${styles.actionBtn} ${statusFilter === f ? styles.actionBtnActive || styles.actionBtn : ''}`}
                style={statusFilter === f ? { borderColor: 'var(--color-primary)', color: 'var(--color-primary)', background: 'var(--primary-muted)' } : {}}
              >
                {f.toUpperCase()}
              </button>
            ))}
          </div>
          
          <button onClick={loadTickets} className={styles.actionBtn}>
            <RefreshCw size={14} />
            <span>Refresh</span>
          </button>
        </div>

        {/* Loading Spinner */}
        {loading ? (
          <div style={{ display: 'flex', justifyContent: 'center', padding: 'var(--space-8)' }}>
            <Loader2 size={32} className="animate-spin" style={{ color: 'var(--color-primary)' }} />
          </div>
        ) : filteredTickets.length === 0 ? (
          <div style={{ textAlign: 'center', padding: 'var(--space-8)', color: 'var(--color-text-muted)' }}>
            No tickets match the selected status filter.
          </div>
        ) : (
          <div className={styles.tableWrapper}>
            <table className={styles.table}>
              <thead>
                <tr>
                  <th className={styles.th}>Ticket Info</th>
                  <th className={styles.th}>Creator</th>
                  <th className={styles.th}>Moderator</th>
                  <th className={styles.th}>Status</th>
                  <th className={styles.th}>Created At</th>
                  <th className={styles.th}>Actions</th>
                </tr>
              </thead>
              <tbody>
                {filteredTickets.map(t => (
                  <tr key={t.id} className={styles.tr}>
                    <td className={styles.td} style={{ fontWeight: 'bold' }}>
                      <div>{t.id}</div>
                      <div style={{ fontSize: '11px', fontWeight: 'normal', color: 'var(--color-text-muted)' }}>
                        {t.categoryName}
                      </div>
                    </td>
                    <td className={styles.td}>{t.creatorUsername}</td>
                    <td className={styles.td}>{t.assignedTo || <span style={{ color: 'var(--color-text-muted)' }}>Unassigned</span>}</td>
                    <td className={styles.td}>
                      <span className={`
                        ${styles.badgeStatus} 
                        ${t.status === 'open' ? styles.badgeOpen : ''}
                        ${t.status === 'claimed' ? styles.badgeClaimed : ''}
                        ${t.status === 'closed' ? styles.badgeClosed : ''}
                      `}>
                        {t.status}
                      </span>
                    </td>
                    <td className={styles.td}>
                      {new Date(t.createdAt).toLocaleDateString()} {new Date(t.createdAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                    </td>
                    <td className={styles.td}>
                      <div className={styles.btnGroup}>
                        {t.status === 'open' && (
                          <button onClick={() => handleClaim(t.id)} className={styles.actionBtn} title="Claim Ticket">
                            <Check size={12} />
                            <span>Claim</span>
                          </button>
                        )}
                        {t.status !== 'closed' ? (
                          <button onClick={() => handleClose(t.id)} className={styles.actionBtn} title="Close Ticket">
                            <Lock size={12} />
                            <span>Close</span>
                          </button>
                        ) : (
                          <>
                            <button onClick={() => handleReopen(t.id)} className={styles.actionBtn} title="Reopen Ticket">
                              <Unlock size={12} />
                              <span>Reopen</span>
                            </button>
                            <button onClick={() => handleDownloadTranscript(t.id)} className={styles.actionBtn} title="Download Transcript">
                              <Download size={12} />
                              <span>Transcript</span>
                            </button>
                            <button onClick={() => handleArchive(t.id)} className={`${styles.actionBtn} ${styles.actionBtnDanger}`} title="Archive / Delete Ticket">
                              <Trash2 size={12} />
                              <span>Archive</span>
                            </button>
                          </>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
