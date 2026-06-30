import styles from "./Skeleton.module.css";

interface SkeletonProps {
  height?: string | number;
  width?: string | number;
  borderRadius?: string;
  className?: string;
  style?: React.CSSProperties;
}

export function Skeleton({ height = "16px", width = "100%", borderRadius, className, style }: SkeletonProps) {
  return (
    <div
      className={`${styles.skeleton} ${className || ""}`}
      style={{ height, width, borderRadius, ...style }}
    />
  );
}

export function SkeletonField() {
  return (
    <div className={styles.field}>
      <Skeleton height="10px" width="60%" />
      <Skeleton height="38px" />
    </div>
  );
}

interface SkeletonCardProps {
  fields?: number;
}

export function SkeletonCard({ fields = 3 }: SkeletonCardProps) {
  return (
    <div className={styles.card}>
      {Array.from({ length: fields }).map((_, i) => (
        <SkeletonField key={i} />
      ))}
    </div>
  );
}

export function SkeletonHeader() {
  return (
    <div className={styles.header}>
      <Skeleton height="24px" width="40%" />
      <Skeleton height="14px" width="65%" />
    </div>
  );
}
