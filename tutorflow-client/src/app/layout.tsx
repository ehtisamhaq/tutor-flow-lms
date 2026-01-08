import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import { Providers } from "@/components/providers";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "TutorFlow - Learn from the Best",
  description:
    "TutorFlow is a modern learning management system that helps students learn from expert instructors through interactive courses, quizzes, and certifications.",
  keywords: [
    "online learning",
    "courses",
    "education",
    "tutorials",
    "certification",
  ],
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={inter.className}>
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
