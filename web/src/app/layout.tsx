import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import Nav from "./nav";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Happy Study - AI 学习与面试准备",
  description: "先考后学，AI 驱动的程序员学习平台",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="zh-CN">
      <body className={`${geistSans.variable} ${geistMono.variable} antialiased min-h-screen`}>
        <div className="h-1 bg-gradient-to-r from-sky-400 via-indigo-500 to-purple-500" />
        <Nav />
        <main className="max-w-4xl mx-auto px-4 py-6">
          {children}
        </main>
        <footer className="border-t border-border mt-12 py-6">
          <div className="max-w-4xl mx-auto px-4 text-center">
            <p className="text-xs text-muted-foreground/60">
              Happy Study · AI 驱动的程序员学习平台
            </p>
          </div>
        </footer>
      </body>
    </html>
  );
}
