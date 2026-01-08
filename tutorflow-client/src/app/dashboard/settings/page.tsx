import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { SettingsForm } from "@/components/dashboard/settings-form";

// Server Component
export default function SettingsPage() {
  return (
    <div className="max-w-2xl space-y-6">
      <div>
        <h2 className="text-2xl font-bold">Settings</h2>
        <p className="text-muted-foreground">
          Manage your account settings and preferences
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Profile</CardTitle>
          <CardDescription>Update your personal information</CardDescription>
        </CardHeader>
        <CardContent>
          <SettingsForm />
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Notifications</CardTitle>
          <CardDescription>
            Configure how you receive notifications
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <label className="flex items-center justify-between">
            <div>
              <p className="font-medium">Email notifications</p>
              <p className="text-sm text-muted-foreground">
                Receive course updates via email
              </p>
            </div>
            <input type="checkbox" className="h-5 w-5 rounded" defaultChecked />
          </label>
          <label className="flex items-center justify-between">
            <div>
              <p className="font-medium">Push notifications</p>
              <p className="text-sm text-muted-foreground">
                Receive browser push notifications
              </p>
            </div>
            <input type="checkbox" className="h-5 w-5 rounded" defaultChecked />
          </label>
          <label className="flex items-center justify-between">
            <div>
              <p className="font-medium">Marketing emails</p>
              <p className="text-sm text-muted-foreground">
                Receive promotional content and offers
              </p>
            </div>
            <input type="checkbox" className="h-5 w-5 rounded" />
          </label>
        </CardContent>
      </Card>
    </div>
  );
}
