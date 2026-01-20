import { GraduationCap } from "lucide-react";
import Link from "next/link";

const footerLinks = [
  {
    title: "Product",
    links: ["Courses", "Learning Paths", "Pricing", "For Business"],
  },
  {
    title: "Company",
    links: ["About", "Careers", "Blog", "Contact"],
  },
  {
    title: "Support",
    links: ["Help Center", "FAQ", "Terms", "Privacy"],
  },
];

export default function Footer() {
  return (
    <footer className="border-t py-12 bg-muted/30">
      <div className="container mx-auto px-4">
        <div className="grid md:grid-cols-4 gap-8">
          <div>
            <Link href="/" className="flex items-center gap-2 mb-4">
              <GraduationCap className="h-6 w-6 text-primary" />
              <span className="font-bold">TutorFlow</span>
            </Link>
            <p className="text-sm text-muted-foreground">
              Empowering learners worldwide with quality education.
            </p>
          </div>

          {footerLinks.map((column) => (
            <div key={column.title}>
              <h4 className="font-semibold mb-4">{column.title}</h4>
              <ul className="space-y-2">
                {column.links.map((link) => (
                  <li key={link}>
                    <Link
                      href="#"
                      className="text-sm text-muted-foreground hover:text-foreground transition-colors"
                    >
                      {link}
                    </Link>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        <div className="border-t mt-8 pt-8 text-center text-sm text-muted-foreground">
          © {new Date().getFullYear()} TutorFlow. All rights reserved.
        </div>
      </div>
    </footer>
  );
}
