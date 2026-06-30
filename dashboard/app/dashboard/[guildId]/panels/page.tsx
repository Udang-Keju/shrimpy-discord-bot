// dashboard/app/dashboard/[guildId]/panels/page.tsx
"use client";

import { useEffect, useRef, useState } from "react";
import { useParams } from "next/navigation";
import {
  Layers,
  Plus,
  Trash2,
  Eye,
  Ticket
} from "lucide-react";
import styles from "@/app/dashboard/[guildId]/dashboard.module.css";
import { ShrimpyAPI, TicketPanel, TicketCategory, DiscordChannel, DiscordRole } from "@/lib/api";
import Dropdown from "@/components/Dropdown";
import { useToast } from "@/hooks/useToast";

const BUTTON_COLORS: Record<string, string> = {
  primary: '#5865F2',
  success: '#3ba55d',
  danger: '#d83c3e',
  secondary: '#4f545c',
};

function colorToHex(n?: number): string {
  if (n === undefined || n === null) return '#5865F2';
  return '#' + Math.max(0, Math.min(0xffffff, n)).toString(16).padStart(6, '0');
}

function hexToColor(hex: string): number {
  return parseInt(hex.replace('#', ''), 16) || 0;
}

// Guards the live preview's <img> tags: an in-progress, non-absolute URL
// (e.g. typed character-by-character) would otherwise resolve relative to
// the current page and fire a real (404) request on every keystroke.
function isImageUrl(url?: string): boolean {
  return !!url && /^https?:\/\//i.test(url);
}

export default function PanelsPage() {
  const params = useParams();
  const guildId = params?.guildId as string;
  const { showToast, updateToast } = useToast();

  const [panels, setPanels] = useState<TicketPanel[]>([]);
  const [channels, setChannels] = useState<DiscordChannel[]>([]);
  const [channelGroups, setChannelGroups] = useState<DiscordChannel[]>([]);
  const [roles, setRoles] = useState<DiscordRole[]>([]);
  const [selectedPanel, setSelectedPanel] = useState<TicketPanel | null>(null);
  const [categories, setCategories] = useState<TicketCategory[]>([]);
  // Working list of role IDs for the panel form (create or edit) — saved as part of the
  // panel payload; the backend reconciles panel_handler_roles to match.
  const [handlerRoleIds, setHandlerRoleIds] = useState<string[]>([]);
  const [selectedHandlerRole, setSelectedHandlerRole] = useState("");
  const [selectedCategory, setSelectedCategory] = useState<TicketCategory | null>(null);
  // Working list of role IDs for the category form — saved as part of the category
  // payload; the backend reconciles category_handler_roles to match.
  const [categoryHandlerRoleIds, setCategoryHandlerRoleIds] = useState<string[]>([]);
  const [selectedCategoryHandlerRole, setSelectedCategoryHandlerRole] = useState("");
  // Guild staff roles (dashboard-access tier) — always invited into tickets, shown
  // read-only in both forms so users see the full handler-role hierarchy.
  const [staffRoleIds, setStaffRoleIds] = useState<string[]>([]);

  // Form state for new panel
  const [newName, setNewName] = useState("Main Support Desk");
  const [newChannelId, setNewChannelId] = useState("");
  const [newPanelStyle, setNewPanelStyle] = useState<'buttons' | 'select_menu'>('buttons');
  const [newContent, setNewContent] = useState("");
  const [newEmbedTitle, setNewEmbedTitle] = useState("Contact Support Services");
  const [newEmbedDesc, setNewEmbedDesc] = useState("Click a button below to open a private ticket.");
  const [newEmbedColor, setNewEmbedColor] = useState<string>('#5865F2');
  const [newAuthorName, setNewAuthorName] = useState("");
  const [newAuthorIconUrl, setNewAuthorIconUrl] = useState("");
  const [newThumbnailUrl, setNewThumbnailUrl] = useState("");
  const [newImageUrl, setNewImageUrl] = useState("");
  const [newFooterText, setNewFooterText] = useState("");
  const [newFooterIconUrl, setNewFooterIconUrl] = useState("");

  // Form state for new category
  const [newCatName, setNewCatName] = useState("");
  const [newCatButtonLabel, setNewCatButtonLabel] = useState("");
  const [newCatButtonStyle, setNewCatButtonStyle] = useState<'primary' | 'secondary' | 'success' | 'danger'>('primary');
  const [newCatEmoji, setNewCatEmoji] = useState("");
  const [newCatDestination, setNewCatDestination] = useState<'thread' | 'channel'>('thread');
  // Thread destination: parent channel the private thread is started from ('' ⇒ panel's channel).
  const [newCatThreadParentId, setNewCatThreadParentId] = useState("");
  // Channel destination: channel group the dedicated channel is placed under ('' ⇒ no group).
  const [newCatChannelGroupId, setNewCatChannelGroupId] = useState("");
  // Template for the opened channel/thread name; supports {user.name}, {user.id}, {category}, {number}.
  const [newCatNameTemplate, setNewCatNameTemplate] = useState('{category}-{number}');
  const [newCatOpenContent, setNewCatOpenContent] = useState("{ping}");
  const [newCatOpenTitle, setNewCatOpenTitle] = useState("");
  const [newCatOpenDesc, setNewCatOpenDesc] = useState("Welcome {mention}!");
  const [newCatOpenColor, setNewCatOpenColor] = useState<string>('#5865F2');
  const [newCatOpenAuthorName, setNewCatOpenAuthorName] = useState("");
  const [newCatOpenAuthorIconUrl, setNewCatOpenAuthorIconUrl] = useState("");
  const [newCatOpenThumbnailUrl, setNewCatOpenThumbnailUrl] = useState("");
  const [newCatOpenImageUrl, setNewCatOpenImageUrl] = useState("");
  const [newCatOpenFooterText, setNewCatOpenFooterText] = useState("");
  const [newCatOpenFooterIconUrl, setNewCatOpenFooterIconUrl] = useState("");

  // Tracks the live selection so async completions can tell whether the
  // panel/category they were issued for is still the one on screen.
  const selectedPanelIdRef = useRef<string | null>(null);
  useEffect(() => {
    selectedPanelIdRef.current = selectedPanel?.id ?? null;
  }, [selectedPanel]);

  const isCreatingPanel = useRef(false);
  const isCreatingCategory = useRef(false);

  const [editingPanelId, setEditingPanelId] = useState<string | null>(null);
  // Whether the Create form is open (distinct from editingPanelId, which means Edit mode).
  const [creatingNew, setCreatingNew] = useState(false);
  const [editingCategoryId, setEditingCategoryId] = useState<string | null>(null);
  // Whether the category Create form is open (distinct from editingCategoryId, which
  // means Edit mode).
  const [creatingNewCategory, setCreatingNewCategory] = useState(false);
  // Holds fields the category form doesn't expose (buttonOrder, ticketNameTemplate,
  // maxTicketsPerUser, autoCloseHours, transcriptChannelId, allowUserClose) so an
  // edit submission doesn't clobber them with create-time defaults.
  const [editingCategoryOriginal, setEditingCategoryOriginal] = useState<TicketCategory | null>(null);

  const resetPanelForm = () => {
    setNewName("Main Support Desk");
    setNewChannelId(channels.length > 0 ? channels[0].id : "");
    setNewContent("");
    setNewEmbedTitle("Contact Support Services");
    setNewEmbedDesc("Click a button below to open a private ticket.");
    setNewEmbedColor('#5865F2');
    setNewAuthorName("");
    setNewAuthorIconUrl("");
    setNewThumbnailUrl("");
    setNewImageUrl("");
    setNewFooterText("");
    setNewFooterIconUrl("");
    setNewPanelStyle('buttons');
    setHandlerRoleIds([]);
  };

  const resetCategoryForm = () => {
    setNewCatName("");
    setNewCatButtonLabel("");
    setNewCatButtonStyle('primary');
    setNewCatEmoji("");
    setNewCatDestination('thread');
    setNewCatThreadParentId("");
    setNewCatChannelGroupId("");
    setNewCatNameTemplate('{category}-{number}');
    setNewCatOpenContent("{ping}");
    setNewCatOpenTitle("");
    setNewCatOpenDesc("Welcome {mention}!");
    setNewCatOpenColor('#5865F2');
    setNewCatOpenAuthorName("");
    setNewCatOpenAuthorIconUrl("");
    setNewCatOpenThumbnailUrl("");
    setNewCatOpenImageUrl("");
    setNewCatOpenFooterText("");
    setNewCatOpenFooterIconUrl("");
    setCategoryHandlerRoleIds([]);
  };

  useEffect(() => {
    async function loadData() {
      try {
        const [panelsData, chansData, groupsData, rolesData, guildConfig] = await Promise.all([
          ShrimpyAPI.listPanels(guildId),
          ShrimpyAPI.getDiscordChannels(guildId),
          ShrimpyAPI.getDiscordChannelGroups(guildId),
          ShrimpyAPI.getDiscordRoles(guildId),
          ShrimpyAPI.getGuildConfig(guildId)
        ]);
        setPanels(panelsData);
        setChannels(chansData);
        setChannelGroups(groupsData);
        setRoles(rolesData);
        setStaffRoleIds(guildConfig.staffRoles);

        if (chansData.length > 0) {
          setNewChannelId(chansData[0].id);
        }
        if (rolesData.length > 0) {
          setSelectedHandlerRole(rolesData[0].id);
          setSelectedCategoryHandlerRole(rolesData[0].id);
        }

        // Start with nothing selected; default to the Create form when panels exist
        // so the user lands on creation rather than an arbitrary panel's edit screen.
        setSelectedPanel(null);
        setSelectedCategory(null);
        setEditingPanelId(null);
        setCreatingNew(panelsData.length > 0);
        setEditingCategoryId(null);
        setEditingCategoryOriginal(null);
      } catch (err) {
        console.error(err);
      }
    }
    loadData();
  }, [guildId]);

  // Switches the selected panel and resets all state scoped to "the panel
  // currently being viewed" in one synchronous step, so the change happens
  // in the event handler that causes it rather than being inferred later in
  // an effect (see https://react.dev/learn/you-might-not-need-an-effect).
  const selectPanel = (p: TicketPanel | null) => {
    setSelectedPanel(p);
    setSelectedCategory(null);
    setEditingCategoryId(null);
    setEditingCategoryOriginal(null);
    setCreatingNewCategory(false);
    resetCategoryForm();
    setEditingPanelId(null);
    resetPanelForm();
    if (!p) {
      setCategories([]);
      setHandlerRoleIds([]);
    }
  };

  useEffect(() => {
    if (selectedPanel) {
      ShrimpyAPI.listCategories(guildId, selectedPanel.id).then(cats => {
        setCategories(cats);
        // Default to the Create form when there's room for another category;
        // at the button-layout cap, leave it closed behind the limit note.
        const atLimit = selectedPanel.panelStyle === 'buttons' && cats.length >= 3;
        setCreatingNewCategory(cats.length > 0 && !atLimit);
      });
      ShrimpyAPI.listPanelHandlerRoles(guildId, selectedPanel.id).then(res => setHandlerRoleIds(res.map(hr => hr.roleId)));
    }
  }, [selectedPanel, guildId]);

  useEffect(() => {
    if (selectedPanel && selectedCategory) {
      ShrimpyAPI.listCategoryHandlerRoles(guildId, selectedPanel.id, selectedCategory.id).then(res => setCategoryHandlerRoleIds(res.map(hr => hr.roleId)));
    } else {
      const timer = setTimeout(() => {
        setCategoryHandlerRoleIds([]);
      }, 0);
      return () => clearTimeout(timer);
    }
  }, [selectedPanel, selectedCategory, guildId]);

  // Clicking a panel row both selects it and opens it for editing in one step,
  // so there's no separate "edit" affordance to hunt for.
  const handleSelectPanelForEdit = (p: TicketPanel) => {
    setCreatingNew(false);
    setSelectedPanel(p);
    setSelectedCategory(null);
    setEditingCategoryId(null);
    setEditingCategoryOriginal(null);
    resetCategoryForm();
    handleEditPanelClick(p);
  };

  // The "+ New Panel" affordance: clear any selection/edit state and open a fresh
  // Create form.
  const handleNewPanelClick = () => {
    selectPanel(null);
    setCreatingNew(true);
  };

  const handleEditPanelClick = (p: TicketPanel) => {
    setEditingPanelId(p.id);
    setNewName(p.name);
    setNewChannelId(p.channelId);
    setNewPanelStyle(p.panelStyle);
    setNewContent(p.content || "");
    setNewEmbedTitle(p.embedTitle || "");
    setNewEmbedDesc(p.embedDescription || "");
    setNewEmbedColor(colorToHex(p.embedColor));
    setNewAuthorName(p.embedMedia?.author?.name || "");
    setNewAuthorIconUrl(p.embedMedia?.author?.iconUrl || "");
    setNewThumbnailUrl(p.embedMedia?.thumbnail?.url || "");
    setNewImageUrl(p.embedMedia?.image?.url || "");
    setNewFooterText(p.embedMedia?.footer?.text || "");
    setNewFooterIconUrl(p.embedMedia?.footer?.iconUrl || "");
  };

  const handleCancelEditPanel = () => {
    setEditingPanelId(null);
    setCreatingNew(false);
    resetPanelForm();
  };

  const handleCreatePanel = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newName.trim() && !newEmbedTitle.trim() && !newEmbedDesc.trim() && !newContent.trim()) {
      showToast("Add at least a title or description to the panel before saving.", "error");
      return;
    }
    if (isCreatingPanel.current) return;
    isCreatingPanel.current = true;

    const hasMedia = !!(newAuthorName || newThumbnailUrl || newImageUrl || newFooterText);
    const payload = {
      channelId: newChannelId,
      name: newName,
      panelStyle: newPanelStyle,
      content: newContent || undefined,
      embedTitle: newEmbedTitle || undefined,
      embedDescription: newEmbedDesc || undefined,
      embedColor: (newEmbedTitle || newEmbedDesc) ? hexToColor(newEmbedColor) : undefined,
      embedMedia: hasMedia ? {
        author: newAuthorName ? { name: newAuthorName, iconUrl: newAuthorIconUrl || undefined } : undefined,
        thumbnail: newThumbnailUrl ? { url: newThumbnailUrl } : undefined,
        image: newImageUrl ? { url: newImageUrl } : undefined,
        footer: newFooterText ? { text: newFooterText, iconUrl: newFooterIconUrl || undefined } : undefined,
      } : undefined,
      handlerRoleIds,
    };

    const editId = editingPanelId;
    isCreatingPanel.current = false;

    if (editId) {
      // Editing: keep the panel open in Edit mode after saving (the form already
      // reflects the edited values), so the user can keep refining it.
      const toastId = showToast(`Updating "${payload.name}"…`, "loading");
      try {
        const p = await ShrimpyAPI.updatePanel(guildId, editId, payload);
        setPanels(prev => prev.map(existing => existing.id === editId ? p : existing));
        setSelectedPanel(prev => prev?.id === editId ? p : prev);
        updateToast(toastId, `"${p.name}" updated.`, "success");
      } catch (err) {
        console.error(err);
        updateToast(toastId, `Failed to update "${payload.name}".`, "error");
      }
      return;
    }

    // Creating: reset the form immediately so the user can start the next panel
    // without waiting for this one's Discord round-trip, staying in Create mode.
    resetPanelForm();

    const toastId = showToast(`Deploying "${payload.name}"…`, "loading");
    try {
      const p = await ShrimpyAPI.createPanel(guildId, payload);
      setPanels(prev => [...prev, p]);
      updateToast(toastId, `"${p.name}" deployed!`, "success");
    } catch (err) {
      console.error(err);
      updateToast(toastId, `Failed to deploy "${payload.name}".`, "error");
    }
  };

  const handleDeletePanel = async (panelId: string) => {
    try {
      await ShrimpyAPI.deletePanel(guildId, panelId);
      const remaining = panels.filter(p => p.id !== panelId);
      setPanels(remaining);
      if (selectedPanel?.id === panelId || editingPanelId === panelId) {
        // Deleting the open panel: fall back to the Create form if others remain,
        // or the empty state (no form) if this was the last one.
        selectPanel(null);
        setCreatingNew(remaining.length > 0);
      }
      showToast("Panel deleted.", "success");
    } catch (err) {
      console.error(err);
      showToast("Failed to delete panel.", "error");
    }
  };

  const handleEditCategoryClick = (c: TicketCategory) => {
    setEditingCategoryId(c.id);
    setEditingCategoryOriginal(c);
    setNewCatName(c.name);
    setNewCatButtonLabel(c.buttonLabel);
    setNewCatButtonStyle(c.buttonStyle);
    setNewCatEmoji(c.emoji || "");
    setNewCatDestination(c.ticketDestination);
    setNewCatThreadParentId(c.threadParentChannelId || "");
    setNewCatChannelGroupId(c.channelCategoryId || "");
    setNewCatNameTemplate(c.ticketNameTemplate || '{category}-{number}');
    setNewCatOpenContent(c.ticketOpenContent || "");
    setNewCatOpenTitle(c.ticketOpenTitle || "");
    setNewCatOpenDesc(c.ticketOpenMessage || "");
    setNewCatOpenColor(colorToHex(c.ticketOpenColor));
    setNewCatOpenAuthorName(c.ticketOpenMedia?.author?.name || "");
    setNewCatOpenAuthorIconUrl(c.ticketOpenMedia?.author?.iconUrl || "");
    setNewCatOpenThumbnailUrl(c.ticketOpenMedia?.thumbnail?.url || "");
    setNewCatOpenImageUrl(c.ticketOpenMedia?.image?.url || "");
    setNewCatOpenFooterText(c.ticketOpenMedia?.footer?.text || "");
    setNewCatOpenFooterIconUrl(c.ticketOpenMedia?.footer?.iconUrl || "");
  };

  // Clicking a category row both selects it (so its handler-roles card shows)
  // and opens it for editing in one step, mirroring handleSelectPanelForEdit.
  const handleSelectCategoryForEdit = (c: TicketCategory) => {
    setCreatingNewCategory(false);
    setSelectedCategory(c);
    handleEditCategoryClick(c);
  };

  // The "+ New Category" affordance: clear any selection/edit state and open
  // a fresh Create form.
  const handleNewCategoryClick = () => {
    setSelectedCategory(null);
    setEditingCategoryId(null);
    setEditingCategoryOriginal(null);
    resetCategoryForm();
    setCreatingNewCategory(true);
  };

  const handleCancelEditCategory = () => {
    setEditingCategoryId(null);
    setEditingCategoryOriginal(null);
    setCreatingNewCategory(false);
    resetCategoryForm();
  };

  const handleCreateCategory = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedPanel || !newCatName || !newCatButtonLabel) return;
    if (!editingCategoryId && selectedPanel.panelStyle === 'buttons' && categories.length >= 3) {
      showToast("Button layout supports up to 3 categories. Switch to Select Menu for more.", "warning");
      return;
    }
    const hasGreeting =
      newCatOpenContent.trim() !== "" ||
      newCatOpenTitle.trim() !== "" ||
      newCatOpenDesc.trim() !== "" ||
      newCatOpenAuthorName.trim() !== "" ||
      newCatOpenThumbnailUrl.trim() !== "" ||
      newCatOpenImageUrl.trim() !== "" ||
      newCatOpenFooterText.trim() !== "";
    if (!hasGreeting) {
      showToast("Add at least a plain text or embed greeting before saving.", "error");
      return;
    }
    if (isCreatingCategory.current) return;
    isCreatingCategory.current = true;

    const panelId = selectedPanel.id;
    const editId = editingCategoryId;
    const original = editingCategoryOriginal;
    const payload = {
      name: newCatName,
      buttonLabel: newCatButtonLabel,
      buttonStyle: newCatButtonStyle,
      emoji: newCatEmoji || undefined,
      buttonOrder: original?.buttonOrder ?? categories.length,
      ticketDestination: newCatDestination,
      threadParentChannelId: newCatDestination === 'thread' ? (newCatThreadParentId || undefined) : undefined,
      channelCategoryId: newCatDestination === 'channel' ? (newCatChannelGroupId || undefined) : undefined,
      ticketNameTemplate: newCatNameTemplate || '{category}-{number}',
      ticketOpenContent: newCatOpenContent || undefined,
      ticketOpenTitle: newCatOpenTitle || undefined,
      ticketOpenMessage: newCatOpenDesc || undefined,
      ticketOpenColor: (newCatOpenTitle || newCatOpenDesc) ? hexToColor(newCatOpenColor) : undefined,
      ticketOpenMedia: (newCatOpenAuthorName || newCatOpenThumbnailUrl || newCatOpenImageUrl || newCatOpenFooterText) ? {
        author: newCatOpenAuthorName ? { name: newCatOpenAuthorName, iconUrl: newCatOpenAuthorIconUrl || undefined } : undefined,
        thumbnail: newCatOpenThumbnailUrl ? { url: newCatOpenThumbnailUrl } : undefined,
        image: newCatOpenImageUrl ? { url: newCatOpenImageUrl } : undefined,
        footer: newCatOpenFooterText ? { text: newCatOpenFooterText, iconUrl: newCatOpenFooterIconUrl || undefined } : undefined,
      } : undefined,
      maxTicketsPerUser: original?.maxTicketsPerUser ?? 1,
      autoCloseHours: original?.autoCloseHours,
      transcriptChannelId: original?.transcriptChannelId,
      allowUserClose: original?.allowUserClose ?? true,
      handlerRoleIds: categoryHandlerRoleIds,
    };

    isCreatingCategory.current = false;

    if (editId) {
      // Editing: keep the category open in Edit mode after saving so the
      // user can keep refining it, matching the panel edit branch.
      const toastId = showToast(`Updating category "${payload.name}"…`, "loading");
      try {
        const c = await ShrimpyAPI.updateCategory(guildId, panelId, editId, payload);
        if (selectedPanelIdRef.current === panelId) {
          setCategories(prev => prev.map(existing => existing.id === editId ? c : existing));
        }
        setSelectedCategory(prev => prev?.id === editId ? c : prev);
        setEditingCategoryOriginal(c);
        updateToast(toastId, `Category "${c.name}" updated.`, "success");
      } catch (err) {
        console.error(err);
        updateToast(toastId, `Failed to update category "${payload.name}".`, "error");
      }
      return;
    }

    // Creating: reset the form immediately so the user can start the next
    // category without waiting for this one's Discord round-trip, staying
    // in Create mode.
    resetCategoryForm();

    const toastId = showToast(`Adding category "${payload.name}"…`, "loading");
    try {
      const c = await ShrimpyAPI.createCategory(guildId, panelId, payload);
      if (selectedPanelIdRef.current === panelId) {
        setCategories(prev => [...prev, c]);
        // Hide the Create form once a button-layout panel hits its 3-category cap.
        if (selectedPanel.panelStyle === 'buttons' && categories.length + 1 >= 3) {
          setCreatingNewCategory(false);
        }
      }
      updateToast(toastId, `Category "${c.name}" added.`, "success");
    } catch (err) {
      console.error(err);
      updateToast(toastId, `Failed to add category "${payload.name}".`, "error");
    }
  };

  const handleDeleteCategory = async (catId: string) => {
    if (!selectedPanel) return;
    const panelId = selectedPanel.id;
    try {
      await ShrimpyAPI.deleteCategory(guildId, panelId, catId);
      const wasEditingDeleted = editingCategoryId === catId;
      if (selectedPanelIdRef.current === panelId) {
        setCategories(prev => prev.filter(c => c.id !== catId));
        if (selectedCategory?.id === catId) {
          setSelectedCategory(null);
        }
        // Dropping below the button-layout cap reopens the Create form,
        // unless another category is currently being edited.
        const underCap = selectedPanel.panelStyle !== 'buttons' || categories.length - 1 < 3;
        if (underCap && (!editingCategoryId || wasEditingDeleted)) {
          setCreatingNewCategory(true);
        }
      }
      if (wasEditingDeleted) {
        setEditingCategoryId(null);
        setEditingCategoryOriginal(null);
        resetCategoryForm();
      }
      showToast("Category deleted.", "success");
    } catch (err) {
      console.error(err);
      showToast("Failed to delete category.", "error");
    }
  };

  // Handler roles are now plain form fields, saved as part of the panel/category
  // payload (the backend reconciles the tables) — these are local list edits only,
  // with no API calls of their own.
  const handleAddHandlerRole = () => {
    if (!selectedHandlerRole) return;
    if (staffRoleIds.includes(selectedHandlerRole)) {
      showToast("Role is already a staff role and is always included.", "warning");
      return;
    }
    if (handlerRoleIds.includes(selectedHandlerRole)) {
      showToast("Role is already a ticket handler for this panel!", "warning");
      return;
    }
    setHandlerRoleIds(prev => [...prev, selectedHandlerRole]);
  };

  const handleRemoveHandlerRole = (roleId: string) => {
    setHandlerRoleIds(prev => prev.filter(id => id !== roleId));
  };

  const handleAddCategoryHandlerRole = () => {
    if (!selectedCategoryHandlerRole) return;
    if (staffRoleIds.includes(selectedCategoryHandlerRole) || handlerRoleIds.includes(selectedCategoryHandlerRole)) {
      showToast("Role is already inherited from staff or panel handler roles.", "warning");
      return;
    }
    if (categoryHandlerRoleIds.includes(selectedCategoryHandlerRole)) {
      showToast("Role is already a ticket handler for this category!", "warning");
      return;
    }
    setCategoryHandlerRoleIds(prev => [...prev, selectedCategoryHandlerRole]);
  };

  const handleRemoveCategoryHandlerRole = (roleId: string) => {
    setCategoryHandlerRoleIds(prev => prev.filter(id => id !== roleId));
  };

  const previewContent = newContent;
  const previewEmbedTitle = newEmbedTitle;
  const previewEmbedDesc = newEmbedDesc;
  const previewEmbedColor = colorToHex(hexToColor(newEmbedColor));
  const previewMedia = (newAuthorName || newThumbnailUrl || newImageUrl || newFooterText) ? {
    author: newAuthorName ? { name: newAuthorName, iconUrl: newAuthorIconUrl || undefined } : undefined,
    thumbnail: newThumbnailUrl ? { url: newThumbnailUrl } : undefined,
    image: newImageUrl ? { url: newImageUrl } : undefined,
    footer: newFooterText ? { text: newFooterText, iconUrl: newFooterIconUrl || undefined } : undefined,
  } : undefined;
  const hasPreviewEmbed = !!(previewEmbedTitle || previewEmbedDesc || previewMedia);
  const atCategoryLimit = !!selectedPanel && selectedPanel.panelStyle === 'buttons' && categories.length >= 3;
  const showCategoryForm = (creatingNewCategory || !!editingCategoryId) && !(!editingCategoryId && atCategoryLimit);
  const previewCategories = selectedPanel ? categories : [];

  // Resolves greeting placeholder tokens with fixed example values for the live preview.
  const catPreviewResolve = (text: string) => text
    .replace(/\{ping\}/g, '@HandlerRole')
    .replace(/\{user\.name\}/g, 'UserName')
    .replace(/\{user\.id\}/g, '123456789')
    .replace(/\{user\}/g, '@UserName')
    .replace(/\{mention\}/g, '@UserName')
    .replace(/\{category\}/g, newCatName || 'Category')
    .replace(/\{id\}/g, 'a1b2c3d4-0000-0000-0000-000000000000');
  const catPreviewContent = catPreviewResolve(newCatOpenContent);
  const catPreviewTitle = catPreviewResolve(newCatOpenTitle);
  const catPreviewDesc = catPreviewResolve(newCatOpenDesc);
  const catPreviewColor = colorToHex(hexToColor(newCatOpenColor));
  const catPreviewMedia = (newCatOpenAuthorName || newCatOpenThumbnailUrl || newCatOpenImageUrl || newCatOpenFooterText) ? {
    author: newCatOpenAuthorName ? { name: catPreviewResolve(newCatOpenAuthorName), iconUrl: newCatOpenAuthorIconUrl || undefined } : undefined,
    thumbnail: newCatOpenThumbnailUrl ? { url: newCatOpenThumbnailUrl } : undefined,
    image: newCatOpenImageUrl ? { url: newCatOpenImageUrl } : undefined,
    footer: newCatOpenFooterText ? { text: catPreviewResolve(newCatOpenFooterText), iconUrl: newCatOpenFooterIconUrl || undefined } : undefined,
  } : undefined;
  const hasCatPreviewEmbed = !!(catPreviewTitle || catPreviewDesc || catPreviewMedia);

  // Handler-role hierarchy: staff roles are always included; panel roles apply to every
  // category on the panel; category roles are additive to just that category. Each role
  // is shown editable in exactly one tier — lower tiers show higher ones read-only.
  const visiblePanelRoleIds = handlerRoleIds.filter(id => !staffRoleIds.includes(id));
  const inheritedCategoryRoleIds = Array.from(new Set([...staffRoleIds, ...handlerRoleIds]));
  const visibleCategoryRoleIds = categoryHandlerRoleIds.filter(id => !inheritedCategoryRoleIds.includes(id));
  const roleName = (roleId: string) => roles.find(r => r.id === roleId)?.name || roleId;

  return (
    <div>
      <div className={styles.sectionHeader}>
        <h2 className={styles.sectionTitle}>Ticket Panels Builder</h2>
        <p className={styles.sectionDesc}>Create interactive ticket creation desks that post plain text and/or an embed to your channels, with one button per category.</p>
      </div>

      <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>

          {/* Active Panels List */}
          <div className={styles.card}>
            {panels.length === 0 ? (
              <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', textAlign: 'center', gap: 'var(--space-3)', padding: 'var(--space-6) var(--space-4)' }}>
                <h3 className={styles.cardTitle}>No ticket panels yet</h3>
                <p className={styles.sectionDesc} style={{ fontSize: '13px', margin: 0 }}>
                  Create one to start collecting tickets from your members.
                </p>
                <button onClick={handleNewPanelClick} className={styles.submitBtn} style={{ marginTop: 'var(--space-2)' }}>
                  <Plus size={16} />
                  <span>New Panel</span>
                </button>
              </div>
            ) : (
              <>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 'var(--space-4)' }}>
                  <div>
                    <h3 className={styles.cardTitle}>Your Ticket Panels</h3>
                    <p className={styles.sectionDesc} style={{ fontSize: '12px', margin: 0 }}>
                      Each panel is a message posted to a channel where members open tickets. Click one to edit it.
                    </p>
                  </div>
                  <button onClick={handleNewPanelClick} className={styles.actionBtn} style={{ flexShrink: 0 }}>
                    <Plus size={14} />
                    <span>New Panel</span>
                  </button>
                </div>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                  {panels.map(p => (
                    <div
                      key={p.id}
                      className={`${styles.actionBtn}`}
                      style={{
                        justifyContent: 'space-between',
                        borderColor: selectedPanel?.id === p.id ? 'var(--color-primary)' : '',
                        background: selectedPanel?.id === p.id ? 'var(--primary-muted)' : '',
                      }}
                      onClick={() => handleSelectPanelForEdit(p)}
                    >
                      <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                        <Layers size={14} style={{ color: 'var(--color-primary)' }} />
                        <span style={{ fontWeight: 'bold' }}>{p.name}</span>
                        <span style={{ fontSize: '11px', color: 'var(--color-text-muted)' }}>in #{channels.find(c => c.id === p.channelId)?.name || p.channelId}</span>
                      </div>
                      <button
                        onClick={(e) => { e.stopPropagation(); handleDeletePanel(p.id); }}
                        style={{ background: 'none', border: 'none', color: 'var(--color-danger)', cursor: 'pointer' }}
                      >
                        <Trash2 size={14} />
                      </button>
                    </div>
                  ))}
                </div>
              </>
            )}
          </div>

          {/* Panel Creator / Editor Form, paired with the Real-time Preview */}
          {(creatingNew || editingPanelId) && (
          <div className={styles.grid} style={{ alignItems: 'start' }}>
          <div className={styles.card}>
            <div>
              <h3 className={styles.cardTitle}>{editingPanelId ? 'Edit Ticket Panel' : 'Create New Ticket Panel'}</h3>
              <p className={styles.sectionDesc} style={{ fontSize: '12px', margin: 0 }}>
                Configure the message members see and the channel it&apos;s posted to.
              </p>
            </div>
            <form onSubmit={handleCreatePanel} style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-4)' }}>
              <div className={styles.fieldGroup}>
                <p className={styles.fieldGroupTitle}>Panel Setup</p>
                <div className={styles.gridHalf} style={{ display: 'grid', gap: 'var(--space-4)' }}>
                  <div className={styles.formGroup}>
                    <label className={styles.label}>Panel Name (internal)</label>
                    <input className={styles.input} type="text" value={newName} onChange={e => setNewName(e.target.value)} required />
                  </div>
                  <div className={styles.formGroup}>
                    <label className={styles.label}>Destination Channel</label>
                    <Dropdown
                      value={newChannelId}
                      onChange={setNewChannelId}
                      options={channels.map(c => ({ value: c.id, label: `#${c.name}` }))}
                    />
                  </div>
                </div>
                <div className={styles.formGroup}>
                  <label className={styles.label}>Button Layout</label>
                  <Dropdown
                    value={newPanelStyle}
                    onChange={val => setNewPanelStyle(val as 'buttons' | 'select_menu')}
                    options={[
                      { value: "buttons", label: "Buttons (up to 3 categories)" },
                      { value: "select_menu", label: "Select Menu (up to 25 categories)" },
                    ]}
                  />
                </div>
              </div>

              <div className={styles.fieldGroup}>
                <p className={styles.fieldGroupTitle}>Handler Roles</p>
                <p className={styles.sectionDesc} style={{ fontSize: '12px', margin: '0 0 8px' }}>
                  Tickets always include your server&apos;s staff roles. Roles added here apply to every category on this panel, in addition to staff roles. This does not grant dashboard access.
                </p>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', marginBottom: '8px' }}>
                  {staffRoleIds.length === 0 && visiblePanelRoleIds.length === 0 && (
                    <div style={{ color: 'var(--color-text-muted)', fontSize: '12px' }}>No handler roles added. Only the ticket opener and the bot will see the channel.</div>
                  )}
                  {staffRoleIds.map(roleId => (
                    <div
                      key={`staff-${roleId}`}
                      className={styles.actionBtn}
                      style={{ justifyContent: 'space-between', cursor: 'default', opacity: 0.7 }}
                      title="Always added — from staff roles, managed on the Roles page."
                    >
                      <span style={{ fontWeight: 'bold' }}>{roleName(roleId)}</span>
                      <span style={{ fontSize: '10px', color: 'var(--color-text-muted)', textTransform: 'uppercase' }}>Staff role</span>
                    </div>
                  ))}
                  {visiblePanelRoleIds.map(roleId => (
                    <div key={roleId} className={styles.actionBtn} style={{ justifyContent: 'space-between' }}>
                      <span style={{ fontWeight: 'bold' }}>{roleName(roleId)}</span>
                      <button
                        type="button"
                        onClick={() => handleRemoveHandlerRole(roleId)}
                        style={{ background: 'none', border: 'none', color: 'var(--color-danger)', cursor: 'pointer' }}
                      >
                        <Trash2 size={12} />
                      </button>
                    </div>
                  ))}
                </div>
                <div style={{ display: 'flex', gap: '8px' }}>
                  <Dropdown
                    value={selectedHandlerRole}
                    onChange={setSelectedHandlerRole}
                    options={roles.filter(r => !staffRoleIds.includes(r.id) && !handlerRoleIds.includes(r.id)).map(r => ({ value: r.id, label: r.name }))}
                    style={{ flex: 1 }}
                  />
                  <button type="button" onClick={handleAddHandlerRole} className={styles.actionBtn} style={{ padding: '0 16px', display: 'flex', alignItems: 'center' }}>
                    <Plus size={14} />
                    <span>Add</span>
                  </button>
                </div>
              </div>

              <div className={styles.fieldGroup}>
                <p className={styles.fieldGroupTitle}>Message</p>
                <p className={styles.sectionDesc} style={{ fontSize: '12px', margin: '0 0 4px' }}>
                  The message posted in the channel. Leave the embed title &amp; description empty to send plain text only.
                </p>

                <div className={styles.formGroup}>
                  <label className={styles.label}>Plain Text Message (optional)</label>
                  <textarea className={styles.textarea} rows={2} value={newContent} onChange={e => setNewContent(e.target.value)} placeholder="Sent as the message's own text, above any embed." />
                </div>

                <div className={styles.formGroup}>
                  <label className={styles.label}>Embed Title</label>
                  <input className={styles.input} type="text" value={newEmbedTitle} onChange={e => setNewEmbedTitle(e.target.value)} />
                </div>

                <div className={styles.formGroup}>
                  <label className={styles.label}>Embed Description</label>
                  <textarea className={styles.textarea} rows={3} value={newEmbedDesc} onChange={e => setNewEmbedDesc(e.target.value)} />
                </div>

                <div className={styles.formGroup}>
                  <label className={styles.label}>Embed Color</label>
                  <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                    <input type="color" value={newEmbedColor} onChange={e => setNewEmbedColor(e.target.value)} style={{ width: '40px', height: '36px', padding: '2px', border: '1px solid var(--color-border)', borderRadius: 'var(--radius-sm)', background: 'none', cursor: 'pointer' }} />
                    <input className={styles.input} type="text" value={newEmbedColor} onChange={e => setNewEmbedColor(e.target.value)} style={{ flex: 1 }} />
                  </div>
                </div>

                <div className={styles.gridHalf} style={{ display: 'grid', gap: 'var(--space-4)' }}>
                  <div className={styles.formGroup}>
                    <label className={styles.label}>Author Name</label>
                    <input className={styles.input} type="text" value={newAuthorName} onChange={e => setNewAuthorName(e.target.value)} />
                  </div>
                  <div className={styles.formGroup}>
                    <label className={styles.label}>Author Icon URL</label>
                    <input className={styles.input} type="text" value={newAuthorIconUrl} onChange={e => setNewAuthorIconUrl(e.target.value)} />
                  </div>
                </div>

                <div className={styles.gridHalf} style={{ display: 'grid', gap: 'var(--space-4)' }}>
                <div className={styles.formGroup}>
                  <label className={styles.label}>Thumbnail URL</label>
                  <input className={styles.input} type="text" value={newThumbnailUrl} onChange={e => setNewThumbnailUrl(e.target.value)} />
                </div>
                <div className={styles.formGroup}>
                  <label className={styles.label}>Main Image URL</label>
                  <input className={styles.input} type="text" value={newImageUrl} onChange={e => setNewImageUrl(e.target.value)} />
                </div>
              </div>

                <div className={styles.gridHalf} style={{ display: 'grid', gap: 'var(--space-4)' }}>
                  <div className={styles.formGroup}>
                    <label className={styles.label}>Footer Text</label>
                    <input className={styles.input} type="text" value={newFooterText} onChange={e => setNewFooterText(e.target.value)} />
                  </div>
                  <div className={styles.formGroup}>
                    <label className={styles.label}>Footer Icon URL</label>
                    <input className={styles.input} type="text" value={newFooterIconUrl} onChange={e => setNewFooterIconUrl(e.target.value)} />
                  </div>
                </div>
              </div>

              <div style={{ display: 'flex', gap: 'var(--space-3)' }}>
                <button type="submit" className={styles.submitBtn} style={{ flex: 1 }}>
                  <Plus size={16} />
                  <span>{editingPanelId ? 'Save Changes' : 'Deploy Panel Desk'}</span>
                </button>
                {(editingPanelId || creatingNew) && (
                  <button type="button" className={styles.actionBtn} onClick={handleCancelEditPanel}>
                    Cancel
                  </button>
                )}
              </div>
            </form>
          </div>

          {/* Real-time Discord Preview Card */}
          <div className={styles.card}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
              <Eye size={16} style={{ color: 'var(--color-accent)' }} />
              <h3 className={styles.cardTitle}>Real-time Discord Preview</h3>
            </div>

            <div style={{ background: '#36393f', border: '1px solid #202225', padding: '16px', borderRadius: '8px', minHeight: 'auto' }}>
              {previewContent && (
                <div style={{ color: '#dcddde', fontSize: '14px', whiteSpace: 'pre-wrap', lineHeight: '1.4', marginBottom: hasPreviewEmbed ? '10px' : 0 }}>
                  {previewContent}
                </div>
              )}

              {hasPreviewEmbed && (
                <div style={{ background: '#2f3136', borderLeft: `4px solid ${previewEmbedColor}`, borderRadius: '4px', padding: '16px', width: '100%', display: 'flex', gap: '12px' }}>
                  <div style={{ flex: 1 }}>
                    {previewMedia?.author?.name && (
                      <div style={{ color: '#ffffff', fontSize: '12px', marginBottom: '8px', display: 'flex', alignItems: 'center', gap: '6px' }}>
                        {isImageUrl(previewMedia.author.iconUrl) && (
                          // eslint-disable-next-line @next/next/no-img-element
                          <img src={previewMedia.author.iconUrl} alt="" style={{ width: '20px', height: '20px', borderRadius: '50%' }} />
                        )}
                        <span>{previewMedia.author.name}</span>
                      </div>
                    )}
                    {previewEmbedTitle && (
                      <div style={{ color: '#ffffff', fontWeight: 'bold', fontSize: '15px', marginBottom: '8px' }}>
                        {previewEmbedTitle}
                      </div>
                    )}
                    {previewEmbedDesc && (
                      <div style={{ color: '#dcddde', fontSize: '13px', whiteSpace: 'pre-wrap', lineHeight: '1.4' }}>
                        {previewEmbedDesc}
                      </div>
                    )}
                    {isImageUrl(previewMedia?.image?.url) && (
                      // eslint-disable-next-line @next/next/no-img-element
                      <img src={previewMedia!.image!.url} alt="" style={{ maxWidth: '100%', borderRadius: '4px', marginTop: '10px' }} />
                    )}
                    {previewMedia?.footer?.text && (
                      <div style={{ color: '#72767d', fontSize: '11px', marginTop: '12px', display: 'flex', alignItems: 'center', gap: '6px' }}>
                        {isImageUrl(previewMedia.footer.iconUrl) && (
                          // eslint-disable-next-line @next/next/no-img-element
                          <img src={previewMedia.footer.iconUrl} alt="" style={{ width: '16px', height: '16px', borderRadius: '50%' }} />
                        )}
                        <span>{previewMedia.footer.text}</span>
                      </div>
                    )}
                  </div>
                  {isImageUrl(previewMedia?.thumbnail?.url) && (
                    // eslint-disable-next-line @next/next/no-img-element
                    <img src={previewMedia!.thumbnail!.url} alt="" style={{ width: '64px', height: '64px', borderRadius: '4px', objectFit: 'cover', flexShrink: 0 }} />
                  )}
                </div>
              )}

              <div style={{ display: 'flex', flexWrap: 'wrap', gap: '8px', marginTop: '14px' }}>
                {previewCategories.length > 0 ? (
                  previewCategories.map(c => (
                    <button
                      key={c.id}
                      style={{
                        backgroundColor: BUTTON_COLORS[c.buttonStyle] || BUTTON_COLORS.primary,
                        color: 'white', border: 'none', padding: '8px 16px', borderRadius: '3px', fontWeight: 500, fontSize: '13px', display: 'flex', alignItems: 'center', gap: '6px'
                      }}
                      disabled
                    >
                      <Ticket size={14} />
                      <span>{c.emoji ? `${c.emoji} ` : ''}{c.buttonLabel}</span>
                    </button>
                  ))
                ) : (
                  <button
                    style={{
                      backgroundColor: BUTTON_COLORS[newCatButtonStyle] || BUTTON_COLORS.primary,
                      color: 'white', border: 'none', padding: '8px 16px', borderRadius: '3px', fontWeight: 500, fontSize: '13px', display: 'flex', alignItems: 'center', gap: '6px', opacity: 0.6
                    }}
                    disabled
                  >
                    <Ticket size={14} />
                    <span>Add a category to see buttons</span>
                  </button>
                )}
              </div>
            </div>
          </div>
          </div>
          )}

          {/* Categories inside Selected Panel */}
          {selectedPanel && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>
            <div className={styles.card}>
              {categories.length === 0 ? (
                <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', textAlign: 'center', gap: 'var(--space-3)', padding: 'var(--space-6) var(--space-4)' }}>
                  <h3 className={styles.cardTitle}>No categories yet</h3>
                  <p className={styles.sectionDesc} style={{ fontSize: '13px', margin: 0 }}>
                    Add one so members have a button to open a ticket.
                  </p>
                  <button onClick={handleNewCategoryClick} className={styles.submitBtn} style={{ marginTop: 'var(--space-2)' }}>
                    <Plus size={16} />
                    <span>New Category</span>
                  </button>
                </div>
              ) : (
                <>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 'var(--space-4)' }}>
                    <div>
                      <h3 className={styles.cardTitle}>Ticket Categories</h3>
                      <p className={styles.sectionDesc} style={{ fontSize: '12px', margin: 0 }}>
                        Each category becomes one button on the panel. Click one to edit it.
                      </p>
                    </div>
                    <button
                      onClick={handleNewCategoryClick}
                      className={styles.actionBtn}
                      style={{ flexShrink: 0 }}
                      disabled={atCategoryLimit}
                      title={atCategoryLimit ? "Button layout supports up to 3 categories." : undefined}
                    >
                      <Plus size={14} />
                      <span>New Category</span>
                    </button>
                  </div>
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', margin: '8px 0' }}>
                    {categories.map(c => (
                      <div
                        key={c.id}
                        className={styles.actionBtn}
                        style={{
                          justifyContent: 'space-between',
                          borderColor: selectedCategory?.id === c.id ? 'var(--color-primary)' : '',
                          background: selectedCategory?.id === c.id ? 'var(--primary-muted)' : '',
                        }}
                        onClick={() => handleSelectCategoryForEdit(c)}
                      >
                        <div>
                          <span style={{ fontWeight: 'bold' }}>{c.emoji ? `${c.emoji} ` : ''}{c.name}</span>
                          <span style={{ fontSize: '11px', color: 'var(--color-text-muted)', marginLeft: '6px' }}>
                            opens a {c.ticketDestination}
                          </span>
                        </div>
                        <button
                          onClick={(e) => { e.stopPropagation(); handleDeleteCategory(c.id); }}
                          style={{ background: 'none', border: 'none', color: 'var(--color-danger)', cursor: 'pointer' }}
                        >
                          <Trash2 size={12} />
                        </button>
                      </div>
                    ))}
                  </div>

                  {atCategoryLimit && (
                    <div style={{ borderTop: '1px solid var(--color-border)', paddingTop: 'var(--space-4)', marginTop: 'var(--space-2)', fontSize: '12px', color: 'var(--color-text-muted)' }}>
                      This panel uses Button layout, which supports up to 3 categories. Delete a category to add another, or create a new panel with Select Menu layout to support more.
                    </div>
                  )}
                </>
              )}
            </div>

          {showCategoryForm && (
          <div className={styles.grid} style={{ alignItems: 'start' }}>
            <div className={styles.card}>
              <div>
                <h3 className={styles.cardTitle}>{editingCategoryId ? 'Edit Category' : 'New Category'}</h3>
                <p className={styles.sectionDesc} style={{ fontSize: '12px', margin: 0 }}>
                  Each category becomes one button on the panel.
                </p>
              </div>
              <form onSubmit={handleCreateCategory} style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-3)' }}>
                <div className={styles.fieldGroup}>
                  <p className={styles.fieldGroupTitle}>Button</p>
                  <div className={styles.formGroup}>
                    <label className={styles.label}>Category Name</label>
                    <input
                      className={styles.input}
                      type="text"
                      placeholder="e.g. Billing Assistance"
                      value={newCatName}
                      onChange={e => setNewCatName(e.target.value)}
                      required
                    />
                  </div>

                  <div className={styles.gridHalf} style={{ display: 'grid', gap: 'var(--space-3)' }}>
                    <div className={styles.formGroup}>
                      <label className={styles.label}>Button Label</label>
                      <input
                        className={styles.input}
                        type="text"
                        placeholder="e.g. Billing Help"
                        value={newCatButtonLabel}
                        onChange={e => setNewCatButtonLabel(e.target.value)}
                        required
                      />
                    </div>
                    <div className={styles.formGroup}>
                      <label className={styles.label}>Button Emoji (optional)</label>
                      <input
                        className={styles.input}
                        type="text"
                        placeholder="🎫"
                        value={newCatEmoji}
                        onChange={e => setNewCatEmoji(e.target.value)}
                      />
                    </div>
                  </div>

                  <div className={styles.formGroup}>
                    <label className={styles.label}>Button Style</label>
                    <Dropdown
                      value={newCatButtonStyle}
                      onChange={val => setNewCatButtonStyle(val as 'primary' | 'secondary' | 'success' | 'danger')}
                      options={[
                        { value: "primary", label: "Primary (Blue)" },
                        { value: "success", label: "Success (Green)" },
                        { value: "danger", label: "Danger (Red)" },
                        { value: "secondary", label: "Secondary (Gray)" },
                      ]}
                    />
                  </div>
                </div>

                <div className={styles.fieldGroup}>
                  <p className={styles.fieldGroupTitle}>Ticket Destination</p>
                  <div className={styles.formGroup}>
                    <label className={styles.label}>Opens As</label>
                    <Dropdown
                      value={newCatDestination}
                      onChange={val => setNewCatDestination(val as 'thread' | 'channel')}
                      options={[
                        { value: "thread", label: "Private Thread" },
                        { value: "channel", label: "Private Channel" },
                      ]}
                    />
                  </div>

                  {newCatDestination === 'thread' ? (
                    <div className={styles.formGroup}>
                      <label className={styles.label}>Thread Parent Channel (optional)</label>
                      <p className={styles.sectionDesc} style={{ fontSize: '12px', margin: '0 0 8px' }}>
                        The channel private threads are started in. Leave as the default to use this panel&apos;s channel.
                      </p>
                      <Dropdown
                        value={newCatThreadParentId}
                        onChange={setNewCatThreadParentId}
                        options={[
                          { value: "", label: "Use panel's channel" },
                          ...channels.map(c => ({ value: c.id, label: `#${c.name}` })),
                        ]}
                      />
                    </div>
                  ) : (
                    <div className={styles.formGroup}>
                      <label className={styles.label}>Channel Group (optional)</label>
                      <p className={styles.sectionDesc} style={{ fontSize: '12px', margin: '0 0 8px' }}>
                        The category the dedicated channel is placed under. Choose &quot;No group&quot; to create it at the server root.
                      </p>
                      <Dropdown
                        value={newCatChannelGroupId}
                        onChange={setNewCatChannelGroupId}
                        options={[
                          { value: "", label: "No group" },
                          ...channelGroups.map(g => ({ value: g.id, label: g.name })),
                        ]}
                      />
                    </div>
                  )}

                  <div className={styles.formGroup}>
                    <label className={styles.label}>Ticket Name</label>
                    <p className={styles.sectionDesc} style={{ fontSize: '12px', margin: '0 0 8px' }}>
                      Name of the opened channel/thread. Placeholders: {'{user.name}'}, {'{user.id}'}, {'{category}'}, {'{number}'}. Discord lowercases and hyphenates dedicated-channel names automatically.
                    </p>
                    <input className={styles.input} type="text" value={newCatNameTemplate} onChange={e => setNewCatNameTemplate(e.target.value)} placeholder="{category}-{number}" />
                  </div>
                </div>

                <div className={styles.fieldGroup}>
                  <p className={styles.fieldGroupTitle}>Handler Roles</p>
                  <p className={styles.sectionDesc} style={{ fontSize: '12px', margin: '0 0 8px' }}>
                    Tickets always include your server&apos;s staff roles and this panel&apos;s handler roles. Roles added here apply only to tickets opened from this category. This does not grant dashboard access.
                  </p>
                  <div style={{ display: 'flex', flexDirection: 'column', gap: '8px', marginBottom: '8px' }}>
                    {staffRoleIds.length === 0 && visiblePanelRoleIds.length === 0 && visibleCategoryRoleIds.length === 0 && (
                      <div style={{ color: 'var(--color-text-muted)', fontSize: '12px' }}>No handler roles added. Only the ticket opener and the bot will see the channel.</div>
                    )}
                    {staffRoleIds.map(roleId => (
                      <div
                        key={`staff-${roleId}`}
                        className={styles.actionBtn}
                        style={{ justifyContent: 'space-between', cursor: 'default', opacity: 0.7 }}
                        title="Always added — from staff roles, managed on the Roles page."
                      >
                        <span style={{ fontWeight: 'bold' }}>{roleName(roleId)}</span>
                        <span style={{ fontSize: '10px', color: 'var(--color-text-muted)', textTransform: 'uppercase' }}>Staff role</span>
                      </div>
                    ))}
                    {visiblePanelRoleIds.map(roleId => (
                      <div
                        key={`panel-${roleId}`}
                        className={styles.actionBtn}
                        style={{ justifyContent: 'space-between', cursor: 'default', opacity: 0.7 }}
                        title="Always added — from this panel's handler roles."
                      >
                        <span style={{ fontWeight: 'bold' }}>{roleName(roleId)}</span>
                        <span style={{ fontSize: '10px', color: 'var(--color-text-muted)', textTransform: 'uppercase' }}>Panel</span>
                      </div>
                    ))}
                    {visibleCategoryRoleIds.map(roleId => (
                      <div key={roleId} className={styles.actionBtn} style={{ justifyContent: 'space-between' }}>
                        <span style={{ fontWeight: 'bold' }}>{roleName(roleId)}</span>
                        <button
                          type="button"
                          onClick={() => handleRemoveCategoryHandlerRole(roleId)}
                          style={{ background: 'none', border: 'none', color: 'var(--color-danger)', cursor: 'pointer' }}
                        >
                          <Trash2 size={12} />
                        </button>
                      </div>
                    ))}
                  </div>
                  <div style={{ display: 'flex', gap: '8px' }}>
                    <Dropdown
                      value={selectedCategoryHandlerRole}
                      onChange={setSelectedCategoryHandlerRole}
                      options={roles.filter(r => !inheritedCategoryRoleIds.includes(r.id) && !categoryHandlerRoleIds.includes(r.id)).map(r => ({ value: r.id, label: r.name }))}
                      style={{ flex: 1 }}
                    />
                    <button type="button" onClick={handleAddCategoryHandlerRole} className={styles.actionBtn} style={{ padding: '0 16px', display: 'flex', alignItems: 'center' }}>
                      <Plus size={14} />
                      <span>Add</span>
                    </button>
                  </div>
                </div>

                <div className={styles.fieldGroup}>
                  <p className={styles.fieldGroupTitle}>Greeting</p>
                  <p className={styles.sectionDesc} style={{ fontSize: '12px', margin: '0 0 4px' }}>
                    Sent inside the opened ticket. Leave the embed title &amp; description empty to send plain text only.
                  </p>

                  <details className={styles.placeholderDropdown}>
                    <summary>Placeholders</summary>
                    <table className={styles.placeholderTable}>
                      <tbody>
                        <tr><td><code>{'{ping}'}</code></td><td>Handler role @mentions. Only notifies staff &amp; auto-adds them to private threads when used in the <strong>plain text</strong> greeting (Discord doesn&apos;t send notifications for mentions inside an embed).</td></tr>
                        <tr><td><code>{'{mention}'}</code></td><td>@mentions the ticket opener</td></tr>
                        <tr><td><code>{'{user.name}'}</code></td><td>Opener&apos;s display name (server nick or username)</td></tr>
                        <tr><td><code>{'{user.id}'}</code></td><td>Opener&apos;s Discord user ID</td></tr>
                        <tr><td><code>{'{category}'}</code></td><td>This category&apos;s name</td></tr>
                        <tr><td><code>{'{id}'}</code></td><td>Ticket&apos;s unique ID</td></tr>
                      </tbody>
                    </table>
                  </details>

                  <div className={styles.formGroup}>
                    <label className={styles.label}>Plain Text Greeting (optional)</label>
                    <textarea className={styles.textarea} rows={2} value={newCatOpenContent} onChange={e => setNewCatOpenContent(e.target.value)} placeholder="e.g. {ping} a new ticket from {mention}" />
                  </div>

                  <div className={styles.formGroup}>
                    <label className={styles.label}>Greeting Embed Title (optional)</label>
                    <input className={styles.input} type="text" value={newCatOpenTitle} onChange={e => setNewCatOpenTitle(e.target.value)} />
                  </div>

                  <div className={styles.formGroup}>
                    <label className={styles.label}>Greeting Embed Description (optional)</label>
                    <textarea className={styles.textarea} rows={2} value={newCatOpenDesc} onChange={e => setNewCatOpenDesc(e.target.value)} />
                  </div>

                  <div className={styles.formGroup}>
                    <label className={styles.label}>Greeting Embed Color</label>
                    <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                      <input type="color" value={newCatOpenColor} onChange={e => setNewCatOpenColor(e.target.value)} style={{ width: '40px', height: '36px', padding: '2px', border: '1px solid var(--color-border)', borderRadius: 'var(--radius-sm)', background: 'none', cursor: 'pointer' }} />
                      <input className={styles.input} type="text" value={newCatOpenColor} onChange={e => setNewCatOpenColor(e.target.value)} style={{ flex: 1 }} />
                    </div>
                  </div>

                  <div className={styles.gridHalf} style={{ display: 'grid', gap: 'var(--space-3)' }}>
                    <div className={styles.formGroup}>
                      <label className={styles.label}>Author Name</label>
                      <input className={styles.input} type="text" value={newCatOpenAuthorName} onChange={e => setNewCatOpenAuthorName(e.target.value)} />
                    </div>
                    <div className={styles.formGroup}>
                      <label className={styles.label}>Author Icon URL</label>
                      <input className={styles.input} type="text" value={newCatOpenAuthorIconUrl} onChange={e => setNewCatOpenAuthorIconUrl(e.target.value)} />
                    </div>
                  </div>

                  <div className={styles.gridHalf} style={{ display: 'grid', gap: 'var(--space-3)' }}>
                    <div className={styles.formGroup}>
                      <label className={styles.label}>Thumbnail URL</label>
                      <input className={styles.input} type="text" value={newCatOpenThumbnailUrl} onChange={e => setNewCatOpenThumbnailUrl(e.target.value)} />
                    </div>
                    <div className={styles.formGroup}>
                      <label className={styles.label}>Main Image URL</label>
                      <input className={styles.input} type="text" value={newCatOpenImageUrl} onChange={e => setNewCatOpenImageUrl(e.target.value)} />
                    </div>
                  </div>

                  <div className={styles.gridHalf} style={{ display: 'grid', gap: 'var(--space-3)' }}>
                    <div className={styles.formGroup}>
                      <label className={styles.label}>Footer Text</label>
                      <input className={styles.input} type="text" value={newCatOpenFooterText} onChange={e => setNewCatOpenFooterText(e.target.value)} />
                    </div>
                    <div className={styles.formGroup}>
                      <label className={styles.label}>Footer Icon URL</label>
                      <input className={styles.input} type="text" value={newCatOpenFooterIconUrl} onChange={e => setNewCatOpenFooterIconUrl(e.target.value)} />
                    </div>
                  </div>
                </div>

                <div style={{ display: 'flex', gap: 'var(--space-3)' }}>
                  <button type="submit" className={styles.submitBtn} style={{ padding: '10px', flex: 1 }}>
                    <Plus size={14} />
                    <span>{editingCategoryId ? 'Save Changes' : 'Add Category'}</span>
                  </button>
                  {(editingCategoryId || creatingNewCategory) && (
                    <button type="button" className={styles.actionBtn} onClick={handleCancelEditCategory}>
                      Cancel
                    </button>
                  )}
                </div>
              </form>
            </div>

            {/* Real-time greeting preview */}
            <div className={styles.card}>
              <div style={{ display: 'flex', alignItems: 'center', gap: '6px' }}>
                <Eye size={16} style={{ color: 'var(--color-accent)' }} />
                <h3 className={styles.cardTitle}>Greeting Preview</h3>
              </div>
              <p className={styles.sectionDesc} style={{ fontSize: '12px', margin: 0 }}>
                Sent inside the opened ticket. Placeholders resolved with example values.
              </p>

              <div style={{ background: '#36393f', border: '1px solid #202225', padding: '16px', borderRadius: '8px' }}>
                {catPreviewContent && (
                  <div style={{ color: '#dcddde', fontSize: '14px', whiteSpace: 'pre-wrap', lineHeight: '1.4', marginBottom: hasCatPreviewEmbed ? '10px' : 0 }}>
                    {catPreviewContent}
                  </div>
                )}

                {hasCatPreviewEmbed && (
                  <div style={{ background: '#2f3136', borderLeft: `4px solid ${catPreviewColor}`, borderRadius: '4px', padding: '16px', display: 'flex', gap: '12px' }}>
                    <div style={{ flex: 1 }}>
                      {catPreviewMedia?.author?.name && (
                        <div style={{ color: '#ffffff', fontSize: '12px', marginBottom: '8px', display: 'flex', alignItems: 'center', gap: '6px' }}>
                          {isImageUrl(catPreviewMedia.author.iconUrl) && (
                            // eslint-disable-next-line @next/next/no-img-element
                            <img src={catPreviewMedia.author.iconUrl} alt="" style={{ width: '20px', height: '20px', borderRadius: '50%' }} />
                          )}
                          <span>{catPreviewMedia.author.name}</span>
                        </div>
                      )}
                      {catPreviewTitle && (
                        <div style={{ color: '#ffffff', fontWeight: 'bold', fontSize: '15px', marginBottom: '8px' }}>
                          {catPreviewTitle}
                        </div>
                      )}
                      {catPreviewDesc && (
                        <div style={{ color: '#dcddde', fontSize: '13px', whiteSpace: 'pre-wrap', lineHeight: '1.4' }}>
                          {catPreviewDesc}
                        </div>
                      )}
                      {isImageUrl(catPreviewMedia?.image?.url) && (
                        // eslint-disable-next-line @next/next/no-img-element
                        <img src={catPreviewMedia!.image!.url} alt="" style={{ maxWidth: '100%', borderRadius: '4px', marginTop: '10px' }} />
                      )}
                      {catPreviewMedia?.footer?.text && (
                        <div style={{ color: '#72767d', fontSize: '11px', marginTop: '12px', display: 'flex', alignItems: 'center', gap: '6px' }}>
                          {isImageUrl(catPreviewMedia.footer.iconUrl) && (
                            // eslint-disable-next-line @next/next/no-img-element
                            <img src={catPreviewMedia!.footer!.iconUrl} alt="" style={{ width: '16px', height: '16px', borderRadius: '50%' }} />
                          )}
                          <span>{catPreviewMedia.footer.text}</span>
                        </div>
                      )}
                    </div>
                    {isImageUrl(catPreviewMedia?.thumbnail?.url) && (
                      // eslint-disable-next-line @next/next/no-img-element
                      <img src={catPreviewMedia!.thumbnail!.url} alt="" style={{ width: '64px', height: '64px', borderRadius: '4px', objectFit: 'cover', flexShrink: 0 }} />
                    )}
                  </div>
                )}

                {(!catPreviewContent && !hasCatPreviewEmbed) && (
                  <div style={{ color: '#72767d', fontSize: '12px', fontStyle: 'italic' }}>
                    Fill in the greeting fields to see a preview here.
                  </div>
                )}
              </div>
            </div>
          </div>
          )}
          </div>
          )}
      </div>
    </div>
  );
}
