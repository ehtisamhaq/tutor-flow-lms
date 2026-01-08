import Link from "next/link";
import { redirect } from "next/navigation";
import { MessageSquare, Send, User } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { authServerFetch } from "@/lib/server-api";

interface Conversation {
  id: string;
  participant: {
    id: string;
    first_name: string;
    last_name: string;
    avatar_url?: string;
  };
  last_message?: {
    content: string;
    created_at: string;
  };
  unread_count: number;
}

// Server Component
export default async function MessagesPage() {
  const conversations = await authServerFetch<Conversation[]>(
    "/messages/conversations"
  );

  if (conversations === null) {
    redirect("/login");
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">Messages</h2>
          <p className="text-muted-foreground">
            Your direct messages with instructors and students
          </p>
        </div>
        <Button>
          <Send className="mr-2 h-4 w-4" />
          New Message
        </Button>
      </div>

      {conversations.length === 0 ? (
        <Card className="text-center py-16">
          <CardContent>
            <MessageSquare className="h-16 w-16 mx-auto mb-4 text-muted-foreground" />
            <h3 className="text-xl font-semibold mb-2">No messages yet</h3>
            <p className="text-muted-foreground mb-6">
              Start a conversation with an instructor or classmate
            </p>
          </CardContent>
        </Card>
      ) : (
        <Card>
          <CardHeader>
            <CardTitle>Conversations</CardTitle>
          </CardHeader>
          <CardContent className="p-0">
            <div className="divide-y">
              {conversations.map((conversation) => (
                <Link
                  key={conversation.id}
                  href={`/dashboard/messages/${conversation.id}`}
                  className="flex items-center gap-4 p-4 hover:bg-muted transition-colors"
                >
                  <div className="h-12 w-12 rounded-full bg-primary/10 flex items-center justify-center shrink-0">
                    {conversation.participant.avatar_url ? (
                      <img
                        src={conversation.participant.avatar_url}
                        alt=""
                        className="h-12 w-12 rounded-full object-cover"
                      />
                    ) : (
                      <User className="h-6 w-6 text-primary" />
                    )}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between">
                      <p className="font-medium truncate">
                        {conversation.participant.first_name}{" "}
                        {conversation.participant.last_name}
                      </p>
                      {conversation.unread_count > 0 && (
                        <span className="bg-primary text-primary-foreground text-xs px-2 py-0.5 rounded-full">
                          {conversation.unread_count}
                        </span>
                      )}
                    </div>
                    {conversation.last_message && (
                      <p className="text-sm text-muted-foreground truncate">
                        {conversation.last_message.content}
                      </p>
                    )}
                  </div>
                </Link>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
