import { revalidatePath } from "next/cache";
import { NextRequest, NextResponse } from "next/server";

export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const { slugs = [], categories = [] } = body;

    // Revalidate affected pages
    revalidatePath("/", "page");
    for (const cat of categories) {
      revalidatePath(`/${cat}`, "page");
    }
    for (const slug of slugs) {
      revalidatePath(`/articles/${slug}`, "page");
    }

    return NextResponse.json({ revalidated: true, timestamp: new Date().toISOString() });
  } catch {
    return NextResponse.json({ error: "Invalid request" }, { status: 400 });
  }
}
