import Link from "next/link";
import { Button } from "@/components/ui/button";
import {
  BookOpen,
  GraduationCap,
  Award,
  Play,
  Star,
  ArrowRight,
  CheckCircle,
} from "lucide-react";

// This is a Server Component by default - no 'use client' needed
export default function HomePage() {
  const stats = [
    { value: "50K+", label: "Students" },
    { value: "200+", label: "Courses" },
    { value: "100+", label: "Instructors" },
    { value: "4.9", label: "Avg Rating" },
  ];

  const features = [
    {
      icon: BookOpen,
      title: "Interactive Courses",
      description:
        "Engage with video lessons, quizzes, and hands-on projects designed by industry experts.",
    },
    {
      icon: Award,
      title: "Verified Certificates",
      description:
        "Earn recognized certificates upon completion to showcase your new skills.",
    },
    {
      icon: Star,
      title: "Learn at Your Pace",
      description:
        "Access courses anytime, anywhere. Learn on your schedule with lifetime access.",
    },
  ];

  return (
    <div className="min-h-screen">
      {/* Hero Section */}
      <section className="relative py-20 md:py-32 bg-gradient-to-br from-primary/5 via-background to-primary/10">
        <div className="container mx-auto px-4">
          <div className="max-w-4xl mx-auto text-center">
            <h1 className="text-4xl md:text-6xl font-bold tracking-tight mb-6">
              Learn from{" "}
              <span className="bg-clip-text text-transparent bg-gradient-to-r from-primary to-purple-600">
                World-Class
              </span>{" "}
              Instructors
            </h1>
            <p className="text-lg md:text-xl text-muted-foreground mb-8 max-w-2xl mx-auto">
              Master new skills with interactive courses, hands-on projects, and
              expert-led content. Start your learning journey today.
            </p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <Button size="lg" asChild>
                <Link href="/courses">
                  Explore Courses
                  <ArrowRight className="ml-2 h-5 w-5" />
                </Link>
              </Button>
              <Button size="lg" variant="outline" asChild>
                <Link href="/register">
                  <Play className="mr-2 h-5 w-5" />
                  Start Free Trial
                </Link>
              </Button>
            </div>

            {/* Stats */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-8 mt-16">
              {stats.map((stat) => (
                <div key={stat.label}>
                  <div className="text-3xl md:text-4xl font-bold text-primary">
                    {stat.value}
                  </div>
                  <div className="text-sm text-muted-foreground">
                    {stat.label}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-20 bg-muted/30">
        <div className="container mx-auto px-4">
          <div className="text-center mb-12">
            <h2 className="text-3xl font-bold mb-4">Why Choose TutorFlow?</h2>
            <p className="text-muted-foreground max-w-2xl mx-auto">
              We provide the best learning experience with cutting-edge features
            </p>
          </div>

          <div className="grid md:grid-cols-3 gap-8">
            {features.map((feature) => (
              <div
                key={feature.title}
                className="bg-card rounded-xl p-6 border shadow-sm hover:shadow-md transition-shadow"
              >
                <div className="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center mb-4">
                  <feature.icon className="h-6 w-6 text-primary" />
                </div>
                <h3 className="text-lg font-semibold mb-2">{feature.title}</h3>
                <p className="text-muted-foreground">{feature.description}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-20">
        <div className="container mx-auto px-4">
          <div className="bg-primary rounded-2xl p-8 md:p-12 text-center text-primary-foreground">
            <h2 className="text-3xl md:text-4xl font-bold mb-4">
              Ready to Start Learning?
            </h2>
            <p className="text-primary-foreground/80 mb-8 max-w-2xl mx-auto">
              Join thousands of students already learning on TutorFlow. Start
              your free trial today.
            </p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <Button
                size="lg"
                variant="secondary"
                className="bg-white text-primary hover:bg-white/90"
                asChild
              >
                <Link href="/register">Get Started for Free</Link>
              </Button>
            </div>
            <div className="flex flex-wrap justify-center gap-4 mt-6 text-sm">
              {[
                "No credit card required",
                "14-day free trial",
                "Cancel anytime",
              ].map((item) => (
                <div key={item} className="flex items-center gap-2">
                  <CheckCircle className="h-4 w-4" />
                  {item}
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>
    </div>
  );
}
