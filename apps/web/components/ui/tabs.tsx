"use client";
import * as React from "react";
import * as TabsPrimitive from "@radix-ui/react-tabs";
import { cn } from "@/lib/utils";
import type { LucideIcon } from "lucide-react";

function Tabs(props: React.ComponentProps<typeof TabsPrimitive.Root>) {
  return <TabsPrimitive.Root data-slot="tabs" {...props} />;
}
function TabsList({ className, ...props }: React.ComponentProps<typeof TabsPrimitive.List>) {
  return <TabsPrimitive.List data-slot="tabs-list" className={cn("glass-tabs", className)} {...props} />;
}
function TabsTrigger({ className, icon: Icon, children, ...props }: React.ComponentProps<typeof TabsPrimitive.Trigger> & {icon?: LucideIcon}) {
  return <TabsPrimitive.Trigger data-slot="tabs-trigger" className={cn("ui-tabs-trigger", className)} {...props}>
    {Icon && <Icon aria-hidden="true" size={17} strokeWidth={1.9}/>}
    {children}
  </TabsPrimitive.Trigger>;
}
function TabsContent({ className, ...props }: React.ComponentProps<typeof TabsPrimitive.Content>) {
  return <TabsPrimitive.Content data-slot="tabs-content" className={cn("ui-tabs-content", className)} {...props}/>;
}
export { Tabs, TabsList, TabsTrigger, TabsContent };
