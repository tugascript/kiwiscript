use leptos::*;

use crate::components::auth::{ExternalAuthButtons, SignUpForm};

#[component]
pub fn SignUpScreen() -> impl IntoView {
    view! {
        <body class="container align-items-center vh-100">
            <div class="row justify-content-center">
                <div class="card">
                    <div class="card-body">
                        <h1 class="card-title mb-3">Sign Up</h1>
                        <SignUpForm />
                        <div class="container">
                            <div class="row justify-content-center">
                                <hr class="my-4 col-11" />
                            </div>
                        </div>
                        <ExternalAuthButtons />
                    </div>
                </div>
            </div>
        </body>
    }
}
