# Shrimpy Admin Dashboard 🦐

This is the Next.js web dashboard frontend for **Shrimpy**. It allows server administrators to configure ticket panels, customize greeting cards, assign reaction roles, and review statistics via an interactive, user-friendly interface.

---

## Technical Details

- **Framework**: Next.js 16 (App Router)
- **Language**: TypeScript
- **Styling**: CSS Modules (Vanilla CSS custom properties defined in [DESIGN_SYSTEM.md](../docs/v1/DESIGN_SYSTEM.md))
- **Icons**: Lucide React
- **Themes**: System preference query with togglable Dark Mode (default) and Light Mode. Includes zero-flash theme persistence in `localStorage`.

---

## Directory Structure

```
dashboard/
├── app/
│   ├── globals.css      # Design System custom properties and root styles
│   ├── layout.tsx       # Main layout container (FOUC prevention script & fonts)
│   ├── page.tsx         # Landing homepage with interactive features
│   └── page.module.css  # Homepage specific styles
├── lib/
│   └── theme.ts         # Theme switching client-side utilities
├── public/              # Static assets
├── Dockerfile.dev       # Development environment container configuration
├── tsconfig.json        # TypeScript configuration
└── package.json         # Scripts and dependencies
```

---

## Environment Variables

To run the dashboard, create a `.env.local` file in this directory (or set them in your deployment host environment):

```env
# Go Backend Endpoint
NEXT_PUBLIC_SHRIMPY_API_URL=http://localhost:8080
```

---

## Local Development

### Prerequisites
Make sure dependencies are installed:
```bash
npm install
```

### Start the development server
```bash
npm run dev
```
Open [http://localhost:3000](http://localhost:3000) in your browser. The page will auto-update as you edit files.

---

## Production Build & Verification

To verify typescript and compilation compatibility, run:
```bash
npm run build
```

This compiles the static and dynamic route optimizations and checks for type errors.

---

## Deployment to Vercel

1. In the Vercel project configuration, set the **Root Directory** to **`dashboard`**.
2. Vercel will automatically configure the **Next.js** build presets.
3. Configure the environment variables in your project settings.
4. Add your Vercel deployment URL callback to your Discord Application redirects:
   ```
   https://<your-vercel-domain>/api/auth/callback/discord
   ```
