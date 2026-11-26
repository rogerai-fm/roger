<script>
  // The one fault line of a signed-in page (it was written out four times). A first read that got no answer
  // (the broker unreachable, a 5xx) cannot tell a signed-out visitor from a signed-in one, so the line never
  // redirects (the login-loop rule in boot.js) and offers both ways on: try again, or sign in and come back.
  import { backHere } from "#lib/account/boot.js";
  let { shown = false, what = "this page", id = "pageError", cls = "fine" } = $props();
</script>
<p class={cls} {id} role="alert" hidden={!shown}>Could not load {what}. <button type="button" class="ghost" onclick={() => location.reload()}>Try again</button> or <a href="/login.html" use:backHere>sign in</a>.</p>
