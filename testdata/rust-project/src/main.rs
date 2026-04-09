use std::collections::HashMap;
use serde::{Serialize, Deserialize};
use crate::config::Settings;

mod config;

pub struct AppState {
    pub db: HashMap<String, String>,
}

pub fn create_app() -> AppState {
    AppState { db: HashMap::new() }
}

fn internal_helper() {}

fn main() {
    let app = create_app();
}
