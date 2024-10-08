use leptos::*;

#[component]
pub fn Navbar() -> impl IntoView {
    view! {
        <nav class="navbar navbar-expand-lg bg-body-tertiary">
            <div class="container-fluid">
                <a class="navbar-brand" href="#">
                    Kiwiscript
                </a>
            </div>
        </nav>
    }
}
