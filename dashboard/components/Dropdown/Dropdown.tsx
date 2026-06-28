// dashboard/components/Dropdown.tsx
"use client";

import { useEffect, useRef, useState } from "react";
import { ChevronDown } from "lucide-react";
import styles from "./Dropdown.module.css";

export interface DropdownOption {
  value: string;
  label: string;
  /** Emoji/text icon, or an image URL to render as an avatar */
  icon?: string;
  /** Optional group header this option is rendered under, in first-seen order */
  group?: string;
}

interface DropdownProps {
  value: string;
  onChange: (value: string) => void;
  options: DropdownOption[];
  placeholder?: string;
  className?: string;
  style?: React.CSSProperties;
}

const isIconUrl = (icon?: string) => !!icon && icon.startsWith("http");

export default function Dropdown({ value, onChange, options, placeholder, className, style }: DropdownProps) {
  const [open, setOpen] = useState(false);
  const wrapperRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (wrapperRef.current && !wrapperRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === "Escape") setOpen(false);
    };
    document.addEventListener("mousedown", handleClickOutside);
    document.addEventListener("keydown", handleEscape);
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
      document.removeEventListener("keydown", handleEscape);
    };
  }, []);

  const selected = options.find(o => o.value === value);

  const groups: { label?: string; options: DropdownOption[] }[] = [];
  for (const opt of options) {
    const last = groups[groups.length - 1];
    if (last && last.label === opt.group) {
      last.options.push(opt);
    } else {
      groups.push({ label: opt.group, options: [opt] });
    }
  }

  const renderIcon = (icon?: string) => {
    if (!icon) return null;
    if (isIconUrl(icon)) {
      // eslint-disable-next-line @next/next/no-img-element
      return <img src={icon} alt="" className={styles.optionIcon} />;
    }
    return <span>{icon}</span>;
  };

  return (
    <div ref={wrapperRef} className={`${styles.wrapper} ${className || ""}`} style={style}>
      <button
        type="button"
        className={`${styles.trigger} ${open ? styles.triggerOpen : ""}`}
        onClick={() => setOpen(o => !o)}
      >
        <span className={styles.triggerLabel}>
          {selected ? (
            <>
              {renderIcon(selected.icon)}
              <span>{selected.label}</span>
            </>
          ) : (
            <span className={styles.triggerPlaceholder}>{placeholder || "Select..."}</span>
          )}
        </span>
        <ChevronDown size={16} className={`${styles.chevron} ${open ? styles.chevronOpen : ""}`} />
      </button>

      {open && (
        <div className={styles.menu}>
          {options.length === 0 ? (
            <div className={styles.empty}>No options available</div>
          ) : (
            groups.map((group, idx) => (
              <div key={group.label ?? `__group_${idx}`}>
                {group.label && <div className={styles.groupLabel}>{group.label}</div>}
                {group.options.map(opt => (
                  <div
                    key={opt.value}
                    className={`${styles.option} ${opt.value === value ? styles.optionActive : ""}`}
                    onClick={() => {
                      onChange(opt.value);
                      setOpen(false);
                    }}
                  >
                    {renderIcon(opt.icon)}
                    <span>{opt.label}</span>
                  </div>
                ))}
              </div>
            ))
          )}
        </div>
      )}
    </div>
  );
}
