import Footer from "../(root)/components/footer";
import Header from "../(root)/components/header";

export default function AuthLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <main>
      <Header />
      {children}
      <Footer />
    </main>
  );
}
