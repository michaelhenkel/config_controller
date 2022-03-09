use crate::resources::traits::{ProcessResource};
use config_client::protos::ssd_git::juniper::net::contrail::cn2::contrail::pkg::apis::core::v1alpha1;

impl ProcessResource for v1alpha1::VirtualRouter {
    fn kind(&self) -> String { "VirtualRouter".to_string() }
    fn add(&self) { 
        println!("add for {} {}/{}",self.kind(), self.metadata.as_ref().unwrap().namespace(), self.metadata.as_ref().unwrap().name())
    }
    fn update(&self) { 
        println!("update for {} {}/{}",self.kind(), self.metadata.as_ref().unwrap().namespace(), self.metadata.as_ref().unwrap().name())
    }
    fn delete(&self) { 
        println!("delete for {} {}/{}",self.kind(), self.metadata.as_ref().unwrap().namespace(), self.metadata.as_ref().unwrap().name())
    }
}