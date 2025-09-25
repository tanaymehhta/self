import { createClient } from '@/lib/supabase/server'
import { NextRequest, NextResponse } from 'next/server'

export async function POST(req: NextRequest) {
  const supabase = createClient()

  // Sign out
  await supabase.auth.signOut()

  return NextResponse.redirect(new URL('/auth/login', req.url), {
    status: 302,
  })
}