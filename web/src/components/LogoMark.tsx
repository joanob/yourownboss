interface LogoMarkProps {
  size?: number;
}

export function LogoMark({ size = 24 }: LogoMarkProps) {
  return (
    <svg viewBox="0 0 28 28" fill="none" style={{ width: size, height: size, display: 'block', flexShrink: 0 }}>
      <rect width="28" height="28" rx="7" fill="#f06449" />
      <rect x="6" y="6" width="7" height="7" rx="1.5" fill="white" />
      <rect x="15" y="6" width="7" height="7" rx="1.5" fill="white" fillOpacity=".6" />
      <rect x="6" y="15" width="7" height="7" rx="1.5" fill="white" fillOpacity=".6" />
      <rect x="15" y="15" width="7" height="7" rx="1.5" fill="white" fillOpacity=".3" />
    </svg>
  );
}
