use once_cell::sync::Lazy;
use reqwest::Client;

static CLIENT: Lazy<Client> = Lazy::new(|| Client::new());

pub fn get_http_client() -> &'static Client {
    &CLIENT
}
