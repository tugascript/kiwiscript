use leptos::*;

// Import the FacebookLoginButton component from its module
use super::FacebookLoginButton;

#[component]
pub fn ExternalAuthButtons() -> impl IntoView {
    view! {
        <div class="external-auth-buttons mt-4 d-flex justify-content-center">
            <FacebookLoginButton />
        </div>
    }
}
