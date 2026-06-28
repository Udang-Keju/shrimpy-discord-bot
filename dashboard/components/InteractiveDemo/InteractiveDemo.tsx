// dashboard/components/InteractiveDemo.tsx
"use client";

import { useState } from "react";
import {
  Ticket,
  UserPlus,
  Tags,
  Check,
  Send,
  MessageSquare,
  Lock
} from "lucide-react";
import styles from "@/app/page.module.css";

interface InteractiveDemoProps {
  activeTab: "tickets" | "welcome" | "roles";
  setActiveTab: (tab: "tickets" | "welcome" | "roles") => void;
}

export default function InteractiveDemo({ activeTab, setActiveTab }: InteractiveDemoProps) {
  // Interactive Ticket State
  const [ticketState, setTicketState] = useState<"none" | "opened" | "claimed" | "closed">("none");
  const [ticketMessages, setTicketMessages] = useState<Array<{author: string, isBot?: boolean, text: string, time: string}>>([]);
  const [inputVal, setInputVal] = useState("");
  const [channelsList, setChannelsList] = useState<Array<{id: string, name: string}>>([
    { id: "rules", name: "rules-and-info" },
    { id: "announcements", name: "announcements" },
    { id: "general", name: "general-chat" }
  ]);
  const [activeChannel, setActiveChannel] = useState("general");

  // Welcome Customizer State
  const [welcomeCardStyle, setWelcomeCardStyle] = useState<"dark" | "light">("dark");
  const [welcomeAvatar, setWelcomeAvatar] = useState("🦐");
  const [welcomeText, setWelcomeText] = useState("Welcome to Shrimpy Server!");

  // Reaction Roles State
  const [assignedRoles, setAssignedRoles] = useState<Record<string, boolean>>([
    { key: "member", active: false },
    { key: "developer", active: false },
    { key: "gamer", active: false }
  ].reduce((acc, curr) => ({ ...acc, [curr.key]: curr.active }), {} as Record<string, boolean>));

  // Simulated Ticket Actions
  const handleOpenTicket = () => {
    if (ticketState !== "none") return;
    setTicketState("opened");
    const newChan = { id: "ticket-001", name: "🎫-ticket-0001" };
    setChannelsList(prev => [...prev, newChan]);
    setActiveChannel("ticket-001");
    setTicketMessages([
      { author: "Shrimpy", isBot: true, text: "👋 Welcome to your support thread. Please wait for assistance.", time: "Today at 11:20 AM" }
    ]);
  };

  const handleClaimTicket = () => {
    if (ticketState !== "opened") return;
    setTicketState("claimed");
    setTicketMessages(prev => [
      ...prev,
      { author: "Shrimpy", isBot: true, text: "🔒 Ticket has been claimed by ModStaff! They will assist you shortly.", time: "Today at 11:21 AM" }
    ]);
  };

  const handleCloseTicket = () => {
    if (ticketState === "closed" || ticketState === "none") return;
    setTicketState("closed");
    setTicketMessages(prev => [
      ...prev,
      { author: "Shrimpy", isBot: true, text: "🚫 Ticket closed. A transcript has been saved to the database.", time: "Today at 11:22 AM" }
    ]);
    setTimeout(() => {
      setChannelsList(prev => prev.filter(c => c.id !== "ticket-001"));
      setActiveChannel("general");
      setTicketState("none");
    }, 4000);
  };

  const handleSendMessage = (e: React.FormEvent) => {
    e.preventDefault();
    if (!inputVal.trim()) return;
    
    const userMsg = { author: "You", text: inputVal.trim(), time: "Today at 11:22 AM" };
    setTicketMessages(prev => [...prev, userMsg]);
    setInputVal("");

    if (ticketState === "opened") {
      setTimeout(() => {
        setTicketMessages(prev => [
          ...prev,
          { author: "Shrimpy", isBot: true, text: "🤖 Shrimpy automations: Need immediate help? Click the 'Claim Ticket' button to notify staff.", time: "Today at 11:22 AM" }
        ]);
      }, 1000);
    }
  };

  // Reaction role toggle
  const toggleRole = (role: string) => {
    setAssignedRoles(prev => ({
      ...prev,
      [role]: !prev[role]
    }));
  };

  return (
    <section id="demo" className={styles.interactive}>
      <div className={styles.sectionHeader}>
        <h2 className={styles.sectionTitle}>See Shrimpy In Action</h2>
        <p className={styles.sectionDesc}>
          Test out Shrimpy&apos;s features right on the page. Switch tabs to see how the dashboard configurations render directly in the Discord chat window.
        </p>
      </div>

      <div className={styles.demoContainer}>
        {/* Sidebar selectors */}
        <div className={styles.demoSidebar}>
          <button 
            onClick={() => setActiveTab("tickets")}
            className={`${styles.demoTabBtn} ${activeTab === "tickets" ? styles.demoTabBtnActive : ""}`}
          >
            <Ticket className={styles.demoTabIcon} size={20} />
            <div>
              <div className={styles.demoTabTitle}>Support Ticketing</div>
              <div className={styles.demoTabDesc}>Spawn support tickets, chat, and close threads.</div>
            </div>
          </button>

          <button 
            onClick={() => setActiveTab("welcome")}
            className={`${styles.demoTabBtn} ${activeTab === "welcome" ? styles.demoTabBtnActive : ""}`}
          >
            <UserPlus className={styles.demoTabIcon} size={20} />
            <div>
              <div className={styles.demoTabTitle}>Welcome Messages</div>
              <div className={styles.demoTabDesc}>Preview welcome templates and edit cards.</div>
            </div>
          </button>

          <button 
            onClick={() => setActiveTab("roles")}
            className={`${styles.demoTabBtn} ${activeTab === "roles" ? styles.demoTabBtnActive : ""}`}
          >
            <Tags className={styles.demoTabIcon} size={20} />
            <div>
              <div className={styles.demoTabTitle}>Reaction Roles</div>
              <div className={styles.demoTabDesc}>Self-assign Discord roles with one-click buttons.</div>
            </div>
          </button>
        </div>

        {/* Right Preview Pane */}
        <div className="demoPreviewPane">
          {activeTab === "tickets" && (
            <div className={styles.discordShell}>
              <div className={styles.discordHeader}>
                <div className={styles.discordHeaderLeft}>
                  <span className={styles.discordHeaderHash}>#</span>
                  <span>{activeChannel === "general" ? "general-chat" : "🎫-ticket-0001"}</span>
                </div>
                <div className={styles.discordHeaderRight}>
                  <MessageSquare size={16} />
                </div>
              </div>

              <div className={styles.discordBody}>
                <div className={styles.discordChannels}>
                  <span className={styles.categoryTitle}>Text Channels</span>
                  <div className={styles.channelList}>
                    {channelsList.map(chan => (
                      <div 
                        key={chan.id} 
                        className={`${styles.channelItem} ${activeChannel === chan.id ? styles.channelItemActive : ""}`}
                        onClick={() => setActiveChannel(chan.id)}
                      >
                        <span className={styles.discordHeaderHash}>#</span>
                        <span>{chan.name}</span>
                      </div>
                    ))}
                  </div>
                </div>

                <div className={styles.discordChat}>
                  <div className={styles.messageArea}>
                    {activeChannel === "general" ? (
                      <>
                        <div className={styles.message}>
                          <div className={`${styles.avatar} ${styles.shrimpyAvatar}`}>🦐</div>
                          <div className={styles.msgContent}>
                            <div className={styles.msgHeader}>
                              <span className={styles.username}>Shrimpy</span>
                              <span className={styles.botTag}>Bot</span>
                              <span className={styles.msgTime}>Today at 10:15 AM</span>
                            </div>
                            <div className={styles.msgText}>
                              Here is the support desk setup. Need help? Open a ticket below.
                            </div>
                            <div className={styles.discordEmbed}>
                              <div className={styles.embedTitle}>Contact Support Services</div>
                              <div className={styles.embedDesc}>Click the button below to open a private ticket. Our staff is available 24/7.</div>
                              <div className={styles.embedFooter}>Response time: &lt; 15 mins</div>
                              <div className={styles.embedButtons}>
                                <button 
                                  className={`${styles.embedBtn} ${styles.embedBtnPrimary}`}
                                  onClick={handleOpenTicket}
                                  disabled={ticketState !== "none"}
                                >
                                  <Ticket size={14} />
                                  <span>Create Ticket</span>
                                </button>
                              </div>
                            </div>
                          </div>
                        </div>
                        {ticketState !== "none" && (
                          <div className={styles.message}>
                            <div className={`${styles.avatar} ${styles.shrimpyAvatar}`}>🦐</div>
                            <div className={styles.msgContent}>
                              <div className={styles.msgHeader}>
                                <span className={styles.username}>Shrimpy</span>
                                <span className={styles.botTag}>Bot</span>
                                <span className={styles.msgTime}>Just now</span>
                              </div>
                              <div className={styles.msgText}>
                                Created private thread <span style={{color: '#7289da', cursor: 'pointer'}} onClick={() => setActiveChannel("ticket-001")}>#🎫-ticket-0001</span> for you.
                              </div>
                            </div>
                          </div>
                        )}
                      </>
                    ) : (
                      <>
                        {ticketMessages.map((m, idx) => (
                          <div key={idx} className={styles.message}>
                            <div className={`${styles.avatar} ${m.isBot ? styles.shrimpyAvatar : ""}`}>
                              {m.isBot ? "🦐" : "U"}
                            </div>
                            <div className={styles.msgContent}>
                              <div className={styles.msgHeader}>
                                <span className={styles.username}>{m.author}</span>
                                {m.isBot && <span className={styles.botTag}>Bot</span>}
                                <span className={styles.msgTime}>{m.time}</span>
                              </div>
                              <div className={styles.msgText}>{m.text}</div>
                              {m.isBot && idx === 0 && (
                                <div className={styles.embedButtons}>
                                  {ticketState === "opened" && (
                                    <button className={`${styles.embedBtn} ${styles.embedBtnSuccess}`} onClick={handleClaimTicket}>
                                      <Check size={14} />
                                      <span>Claim Ticket</span>
                                    </button>
                                  )}
                                  {ticketState !== "closed" && (
                                    <button className={`${styles.embedBtn}`} onClick={handleCloseTicket}>
                                      <Lock size={14} />
                                      <span>Close Ticket</span>
                                    </button>
                                  )}
                                </div>
                              )}
                            </div>
                          </div>
                        ))}
                      </>
                    )}
                  </div>

                  <form className={styles.chatInputArea} onSubmit={handleSendMessage}>
                    <div className={styles.chatInputWrapper}>
                      <input 
                        type="text" 
                        placeholder={activeChannel === "general" ? "You cannot type in this channel" : "Type a message in the ticket..."} 
                        className={styles.chatInput} 
                        value={inputVal}
                        onChange={(e) => setInputVal(e.target.value)}
                        disabled={activeChannel === "general" || ticketState === "closed"}
                      />
                      <button type="submit" style={{background: 'none', border: 'none', color: '#b9bbbe', cursor: 'pointer'}}>
                        <Send size={16} />
                      </button>
                    </div>
                  </form>
                </div>
              </div>
            </div>
          )}

          {activeTab === "welcome" && (
            <div style={{display: 'flex', flexDirection: 'column', gap: '20px'}}>
              <div className={`${styles.welcomeCard} ${welcomeCardStyle === "light" ? styles.welcomeCardLight : ""}`}>
                <div className={styles.cardBlob}></div>
                <div className={styles.welcomeUserAvatar}>{welcomeAvatar}</div>
                <h3 className={styles.welcomeTitle}>{welcomeText}</h3>
                <p className={styles.welcomeDesc}>We are excited to have you here! Feel free to assign yourself some roles.</p>
                <div className={styles.welcomeMeta}>
                  <span>ID: #99318</span>
                  <span>•</span>
                  <span>Server: Shrimpy Sandbox</span>
                </div>
              </div>

              <div className={styles.customizerControls}>
                <div style={{display: 'flex', flexDirection: 'column', gap: '5px'}}>
                  <span style={{fontSize: '12px', color: 'var(--color-text-muted)', fontWeight: 'bold'}}>CARD STYLE</span>
                  <div style={{display: 'flex', gap: '5px'}}>
                    <button 
                      className={`${styles.ctrlBtn} ${welcomeCardStyle === "dark" ? styles.ctrlBtnActive : ""}`}
                      onClick={() => setWelcomeCardStyle("dark")}
                    >
                      Dark Theme
                    </button>
                    <button 
                      className={`${styles.ctrlBtn} ${welcomeCardStyle === "light" ? styles.ctrlBtnActive : ""}`}
                      onClick={() => setWelcomeCardStyle("light")}
                    >
                      Light Theme
                    </button>
                  </div>
                </div>

                <div style={{display: 'flex', flexDirection: 'column', gap: '5px'}}>
                  <span style={{fontSize: '12px', color: 'var(--color-text-muted)', fontWeight: 'bold'}}>SELECT AVATAR EMOJI</span>
                  <div style={{display: 'flex', gap: '5px'}}>
                    {["🦐", "🐠", "🌊", "🌴"].map(em => (
                      <button 
                        key={em}
                        className={`${styles.ctrlBtn} ${welcomeAvatar === em ? styles.ctrlBtnActive : ""}`}
                        onClick={() => setWelcomeAvatar(em)}
                        style={{padding: '8px 12px'}}
                      >
                        {em}
                      </button>
                    ))}
                  </div>
                </div>

                <div style={{display: 'flex', flexDirection: 'column', gap: '5px', gridColumn: 'span 2'}}>
                  <span style={{fontSize: '12px', color: 'var(--color-text-muted)', fontWeight: 'bold'}}>EDIT TEXT HEADER</span>
                  <input 
                    type="text" 
                    className={styles.ctrlBtn}
                    style={{background: 'var(--bg-surface-elevated)', border: '1px solid var(--color-border)', textAlign: 'left'}}
                    value={welcomeText}
                    onChange={(e) => setWelcomeText(e.target.value)}
                  />
                </div>
              </div>
            </div>
          )}

          {activeTab === "roles" && (
            <div className={styles.reactionWidget}>
              <div className={styles.reactionMsg}>
                <h3 className={styles.reactionHeading}>Roles Picker Desk</h3>
                <p className={styles.reactionDesc}>Click on the reaction badges below to self-assign tags in this guild. Hover to see active allocations.</p>
              </div>

              <div style={{display: 'flex', flexDirection: 'column', gap: '15px'}}>
                <div className={styles.reactionItem}>
                  <div className={styles.roleLabel}>
                    <span>🦐</span>
                    <span>Server Member</span>
                    {assignedRoles.member && <span className={styles.roleBadge}>Assigned</span>}
                  </div>
                  <button 
                    className={`${styles.reactionBtn} ${assignedRoles.member ? styles.reactionBtnActive : ""}`}
                    onClick={() => toggleRole("member")}
                  >
                    {assignedRoles.member ? <Check size={14} /> : null}
                    <span>React with 🦐</span>
                  </button>
                </div>

                <div className={styles.reactionItem}>
                  <div className={styles.roleLabel}>
                    <span>🛠️</span>
                    <span>Guild Developer</span>
                    {assignedRoles.developer && <span className={styles.roleBadge}>Assigned</span>}
                  </div>
                  <button 
                    className={`${styles.reactionBtn} ${assignedRoles.developer ? styles.reactionBtnActive : ""}`}
                    onClick={() => toggleRole("developer")}
                  >
                    {assignedRoles.developer ? <Check size={14} /> : null}
                    <span>React with 🛠️</span>
                  </button>
                </div>

                <div className={styles.reactionItem}>
                  <div className={styles.roleLabel}>
                    <span>🎮</span>
                    <span>Community Gamer</span>
                    {assignedRoles.gamer && <span className={styles.roleBadge}>Assigned</span>}
                  </div>
                  <button 
                    className={`${styles.reactionBtn} ${assignedRoles.gamer ? styles.reactionBtnActive : ""}`}
                    onClick={() => toggleRole("gamer")}
                  >
                    {assignedRoles.gamer ? <Check size={14} /> : null}
                    <span>React with 🎮</span>
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </section>
  );
}
