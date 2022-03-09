pub mod k8s {
    pub mod io {
        pub mod api {
            pub mod core {
                pub mod v1 {
                    include!("k8s.io.api.core.v1.rs");
                }
            }
        }
        pub mod apimachinery {
            pub mod pkg {
                pub mod api {
                    pub mod resource {
                        include!("k8s.io.apimachinery.pkg.api.resource.rs");
                    }
                }
                pub mod apis {
                    pub mod meta {
                        pub mod v1 {
                            include!("k8s.io.apimachinery.pkg.apis.meta.v1.rs");
                        }
                    }
                }
                pub mod util {
                    pub mod intstr {
                        include!("k8s.io.apimachinery.pkg.util.intstr.rs");
                    }
                }
                pub mod runtime {
                    pub mod schema {
                        include!("k8s.io.apimachinery.pkg.runtime.schema.rs");
                    }
                    include!("k8s.io.apimachinery.pkg.runtime.rs");
                }
            }
        }
    }
}
pub mod ssd_git {
    pub mod juniper {
        pub mod net {
            pub mod contrail {
                pub mod cn2 {
                    pub mod contrail {
                        pub mod pkg {
                            pub mod apis {
                                pub mod core {
                                    pub mod v1alpha1 {
                                        include!("ssd_git.juniper.net.contrail.cn2.contrail.pkg.apis.core.v1alpha1.rs");
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}
pub mod github {
    pub mod com {
        pub mod michaelhenkel {
            pub mod config_controller {
                pub mod pkg {
                    pub mod apis {
                        pub mod v1 {
                            include!("github.com.michaelhenkel.config_controller.pkg.apis.v1.rs");
                        }
                    }
                }
            }
        }
    }
}
pub mod sigs {
    pub mod k8s {
        pub mod io {
            pub mod apiserver_builder_alpha {
                pub mod pkg {
                    pub mod builders {
                        include!("sigs.k8s.io.apiserver_builder_alpha.pkg.builders.rs");
                    }
                }
            }
        }
    }
}
