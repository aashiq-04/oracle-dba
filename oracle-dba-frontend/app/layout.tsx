'use client';

import './globals.css';
import { ApolloProvider } from '@apollo/client/react';
import apolloClient from '@/lib/apollo-client';

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <head>
        <title>Oracle DBA Platform</title>
        <meta name="description" content="Enterprise Oracle Database Monitoring Platform" />
      </head>
      <body className="bg-gray-50">
        <ApolloProvider client={apolloClient}>
          {children}
        </ApolloProvider>
      </body>
    </html>
  );
}