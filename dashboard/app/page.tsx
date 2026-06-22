// dashboard/app/page.tsx
"use client";

import { useEffect, useState } from "react";
import { 
  Ticket, 
  UserPlus, 
  Tags, 
  Settings, 
  Sun, 
  Moon, 
  Sparkles, 
  ChevronRight, 
  ShieldAlert, 
  Check, 
  Send,
  MessageSquare,
  Lock,
  Layers,
  Heart
} from "lucide-react";
import styles from "./page.module.css";
import { getSavedTheme, applyTheme, Theme } from "../lib/theme";

export default function Home() {
  const [mounted, setMounted] = useState(false);
  const [theme, setTheme] = useState<Theme>("dark");
  const [activeTab, setActiveTab] = useState<"tickets" | "welcome" | "roles">("tickets");
  
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
  const [assignedRoles, setAssignedRoles] = useState<Record<string, boolean>>({
    member: false,
    developer: false,
    gamer: false
  });

  // Handle Theme switching
  useEffect(() => {
    setMounted(true);
    setTheme(getSavedTheme());
  }, []);

  const toggleTheme = () => {
    const nextTheme = theme === "dark" ? "light" : "dark";
    setTheme(nextTheme);
    applyTheme(nextTheme);
  };

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

  if (!mounted) {
    return null;
  }

  return (
    <div className={styles.wrapper}>
      <div className={styles.gridOverlay}></div>

      {/* HEADER NAVBAR */}
      <header className={`${styles.header} glass-effect`}>
        <div className={styles.headerInner}>
          <div className={styles.brand}>
            <span className={styles.brandIcon}>🦐</span>
            <span>Shrimpy</span>
          </div>

          <nav className={styles.nav}>
            <a href="#features" className={styles.navLink}>Features</a>
            <a href="#demo" className={styles.navLink}>Interactive Demo</a>
            <a href="#about" className={styles.navLink}>About</a>
            <div className={styles.navActions}>
              <button 
                onClick={toggleTheme} 
                className={styles.themeBtn}
                title={`Switch to ${theme === 'dark' ? 'Light' : 'Dark'} Mode`}
                aria-label="Toggle Theme"
              >
                {theme === "dark" ? <Sun size={18} /> : <Moon size={18} />}
              </button>
              <a 
                href={`${process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"}/api/v1/auth/login`}
                className={`${styles.btn} ${styles.discordBtn}`}
              >
                <Sparkles size={16} />
                <span>Login with Discord</span>
              </a>
            </div>
          </nav>
        </div>
      </header>

      {/* HERO SECTION */}
      <section className={`${styles.hero} hero-gradient`}>
        <div className={styles.heroInner}>
          <div className={styles.heroContent}>
            <div className={styles.badge}>
              <span>🤖 Multi-Bot Configurable Dashboard</span>
            </div>
            <h1 className={styles.title}>
              Manage Your Guilds with <span className="gradient-text">Shrimpy</span>
            </h1>
            <p className={styles.subtitle}>
              An ultra-modern, high-performance Discord companion offering dynamic support ticket panels, customizable greetings, and responsive reaction roles.
            </p>
            <div className={styles.actions}>
              <button className={`${styles.btn} ${styles.btnPrimary}`}>
                <span>Add to Discord</span>
                <ChevronRight size={16} />
              </button>
              <a href="#demo" className={`${styles.btn} ${styles.btnSecondary}`}>
                <span>Try Interactive Demo</span>
              </a>
            </div>

            <div className={styles.statsRow}>
              <div className={styles.statItem}>
                <span className={styles.statVal}>99.9%</span>
                <span className={styles.statLbl}>Uptime</span>
              </div>
              <div className={styles.statItem}>
                <span className={styles.statVal}>&lt; 50ms</span>
                <span className={styles.statLbl}>Response Time</span>
              </div>
              <div className={styles.statItem}>
                <span className={styles.statVal}>100k+</span>
                <span className={styles.statLbl}>Members Assisted</span>
              </div>
            </div>
          </div>

          <div className={styles.heroVisual}>
            <div className={styles.glowBlob}></div>
            {/* Displaying simple mock widget to make it beautiful */}
            <div className={styles.welcomeCard}>
              <div className={styles.cardBlob}></div>
              <div className={styles.welcomeUserAvatar}>🦐</div>
              <h3 className={styles.welcomeTitle}>Welcome, New Member!</h3>
              <p className={styles.welcomeDesc}>Say hi in #general and get started!</p>
              <div className={styles.welcomeMeta}>
                <span>ID: #89139</span>
                <span>•</span>
                <span>Member count: 1,248</span>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* FEATURES SECTION */}
      <section id="features" className={styles.features}>
        <div className={styles.sectionHeader}>
          <h2 className={styles.sectionTitle}>Everything You Need for Community Management</h2>
          <p className={styles.sectionDesc}>
            Ditch complex setup commands. Configure support pipelines, interactive role assignment panels, and custom announcements directly in our web-dashboard.
          </p>
        </div>

        <div className={styles.featuresGrid}>
          <div className={styles.card} onClick={() => setActiveTab("tickets")}>
            <div className={styles.iconWrapper}>
              <Ticket size={24} />
            </div>
            <h3 className={styles.cardTitle}>Support Ticket Panels</h3>
            <p className={styles.cardDesc}>
              Create rich, button-based ticket builders. When members click to get help, Shrimpy spawns private, temporary threads with custom greeting overlays.
            </p>
          </div>

          <div className={styles.card} onClick={() => setActiveTab("welcome")}>
            <div className={styles.iconWrapper}>
              <UserPlus size={24} />
            </div>
            <h3 className={styles.cardTitle}>Custom Welcome Messages</h3>
            <p className={styles.cardDesc}>
              Set up personalized direct messages and public greetings to welcome new users, automatically applying roles on server join.
            </p>
          </div>

          <div className={styles.card} onClick={() => setActiveTab("roles")}>
            <div className={styles.iconWrapper}>
              <Tags size={24} />
            </div>
            <h3 className={styles.cardTitle}>Interactive Reaction Roles</h3>
            <p className={styles.cardDesc}>
              Enable self-assignable roles. Users simply click message reaction buttons to toggle server tags, avoiding manual staff moderation.
            </p>
          </div>

          <div className={styles.card}>
            <div className={styles.iconWrapper}>
              <Settings size={24} />
            </div>
            <h3 className={styles.cardTitle}>Multi-Bot Registration</h3>
            <p className={styles.cardDesc}>
              Hosting multiple configurations? Manage all your custom Discord applications under a single GORM relational backend in pgxpool.
            </p>
          </div>
        </div>
      </section>

      {/* INTERACTIVE DEMO */}
      <section id="demo" className={styles.interactive}>
        <div className={styles.sectionHeader}>
          <h2 className={styles.sectionTitle}>See Shrimpy In Action</h2>
          <p className={styles.sectionDesc}>
            Test out Shrimpy's features right on the page. Switch tabs to see how the dashboard configurations render directly in the Discord chat window.
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

      {/* FOOTER SECTION */}
      <footer className={styles.footer}>
        <div className={styles.footerInner}>
          <div className={styles.footerBrand}>
            <div className={styles.brand}>
              <span>🦐</span>
              <span>Shrimpy</span>
            </div>
            <p className={styles.footerDesc}>
              A highly performant Discord utility system engineered in Go with Next.js dashboards.
            </p>
          </div>

          <div className={styles.footerCol}>
            <span className={styles.footerTitle}>Resources</span>
            <ul className={styles.footerLinks}>
              <li><a href="#" className={styles.footerLink}>Documentation</a></li>
              <li><a href="#" className={styles.footerLink}>Commands list</a></li>
              <li><a href="#" className={styles.footerLink}>API Guides</a></li>
            </ul>
          </div>

          <div className={styles.footerCol}>
            <span className={styles.footerTitle}>Support</span>
            <ul className={styles.footerLinks}>
              <li><a href="#" className={styles.footerLink}>Join Discord</a></li>
              <li><a href="#" className={styles.footerLink}>Report Bug</a></li>
              <li><a href="#" className={styles.footerLink}>Status Page</a></li>
            </ul>
          </div>

          <div className={styles.footerCol}>
            <span className={styles.footerTitle}>Legal</span>
            <ul className={styles.footerLinks}>
              <li><a href="#" className={styles.footerLink}>Privacy Policy</a></li>
              <li><a href="#" className={styles.footerLink}>Terms of Service</a></li>
            </ul>
          </div>
        </div>

        <div className={styles.footerBottom}>
          <span>&copy; {new Date().getFullYear()} Shrimpy Bot. All rights reserved.</span>
          <span style={{display: 'flex', alignItems: 'center', gap: '4px'}}>
            Made with <Heart size={12} style={{color: 'var(--color-primary)'}} /> by the Engineering Team
          </span>
        </div>
      </footer>
    </div>
  );
}
