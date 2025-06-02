import "@typespec/http";

using Http;
@service(#{ title: "Property Management Service" })
@server("http://localhost:8080", "Local Development Server")
@server("https://property-management.fly.dev", "Production Server")

namespace PropertyManagementService;

model PostalAddress {
  address_line_1: string;
  address_line_2?: string;
  suburb: string;
  postcode: string;
  state: string;
  country: string;
}

model OptionalPostalAddress {
  address_line_1?: string;
  address_line_2?: string;
  suburb?: string;
  postcode?: string;
  state?: string;
  country?: string;
}

@doc("Allows granular sorting by street name and number")
model StructuredAddress {
  street_number: string;
  street_name: string;
  suburb: string;
  postcode: string;
  state: string;
  country: string;
}

model OptionalStructuredAddress {
  street_number?: string;
  street_name?: string;
  suburb?: string;
  postcode?: string;
  state?: string;
  country?: string;
}

model Landlord {
    @visibility(Lifecycle.Read)
    @format("uuid")
    id: string;
    name: string;
    @format("email")
    email: string;
    mobile: string;
    phone?: string;
    ...PostalAddress;
    is_archived?: offsetDateTime;
    created_at: offsetDateTime;
    updated_at: offsetDateTime;
}

model CreateLandlord {
  name: string;
  @format("email")
  email: string;
  mobile: string;
  phone?: string;
  ...PostalAddress;
}

model UpdateLandlord {
    name?: string;
    @format("email")
    email?: string;
    mobile?: string;
    phone?: string;
    ...OptionalPostalAddress;
    is_archived?: offsetDateTime | null;
}

model LandlordList {
  items: Landlord[];
  pagination: PaginatedMetadata;
}

model Property {
    @visibility(Lifecycle.Read)
    @format("uuid")
    id: string;
    @format("uuid")
    landlord_id: string;
    ...StructuredAddress;
    management_fee: float64;
    management_gained: plainDate;
    management_lost?: plainDate;
    is_archived?: offsetDateTime;
    created_at: offsetDateTime;
    updated_at: offsetDateTime;
}

model CreateProperty {
  @format("uuid")
  landlord_id: string;
  ...StructuredAddress;
  management_fee: float64;
  management_gained: plainDate;
}

model UpdateProperty {
  ...OptionalStructuredAddress;
  management_fee?: float64;
  management_gained?: plainDate;
  management_lost?: plainDate | null;
  is_archived?: offsetDateTime | null;
}

model PropertyList {
  items: Property[];
  pagination: PaginatedMetadata;
}

model Tenant {
    @visibility(Lifecycle.Read)
    @format("uuid")
    id: string;
    @format("uuid")
    property_id: string;
    name: string;
    @format("email")
    email: string;
    mobile: string;
    phone?: string;
    paid_from: plainDate;
    paid_to: plainDate;
    rental_amount: float64;
    frequency: string;
    original_start_date: plainDate;
    start_date: plainDate;
    end_date: plainDate;
    vacate_date?: plainDate;
    termination_date?: plainDate;
    termination_reason?: string;
    is_archived?: offsetDateTime;
    created_at: offsetDateTime;
    updated_at: offsetDateTime;
} 

model CreateTenant {
  @format("uuid")
  property_id: string;
  name: string;
  @format("email")
  email: string;
  mobile: string;
  phone?: string;
  paid_to: plainDate;
  rental_amount: float64;
  frequency: string;
  original_start_date: plainDate;
  start_date: plainDate;
  end_date: plainDate;
}

model UpdateTenant {
  name?: string;
  @format("email")
  email?: string;
  mobile?: string;
  phone?: string;
  paid_from?: plainDate;
  paid_to?: plainDate;
  rental_amount?: float64;
  frequency?: string;
  original_start_date?: plainDate;
  start_date?: plainDate;
  end_date?: plainDate;
  vacate_date?: plainDate | null;
  termination_date?: plainDate | null;
  termination_reason?: string | null;
  is_archived?: offsetDateTime | null;
}

model TenantList {
  items: Tenant[];
  pagination: PaginatedMetadata;
}

@error
model Error {
  code: int32;
  message: string;
}

model PaginatedMetadata {
  total: int32;
  count: int32;
  per_page: int32;
  current_page: int32;
  total_pages: int32;
}

@route("/landlords")
namespace Landlords {
  @useAuth(BearerAuth)
  @tag("Landlord")
  @get
  op list(@query page?: int32, @query limit?: int32, @query name?: string, @query archived_only?: boolean): {
    @statusCode statusCode: 200;
    @body landlords: LandlordList;
  } | {
    @statusCode statusCode: 500;
    @body error: Error
  };

  @useAuth(BearerAuth)
  @tag("Landlord")
  @get
  op get(@path id: string): {
    @statusCode statusCode: 200;
    @body landlord: Landlord;
  } | {
    @statusCode statusCode: 401;
    @body error: Error;
  } | {
    @statusCode statusCode: 404;
    @body error: Error;
  } | {
    @statusCode statusCode: 500;
    @body error: Error;
  };

  @useAuth(BearerAuth)
  @tag("Landlord")
  @post
  op create(@body landlord: CreateLandlord): {
    @statusCode statusCode: 201;
    @body landlord: Landlord;
  } | {
    @statusCode statusCode: 400;
    @body error: Error;
  } | {
    @statusCode statusCode: 500;
    @body error: Error;
  };

  @useAuth(BearerAuth)
  @tag("Landlord")
  @patch
  op update(@path id: string, @body landlord: UpdateLandlord): {
    @statusCode statusCode: 200;
    @body landlord: Landlord;
  } | {
    @statusCode statusCode: 400;
    @body error: Error;
  } | {
    @statusCode statusCode: 401;
    @body error: Error;
  } | {
    @statusCode statusCode: 404;
    @body error: Error;
  } | {
    @statusCode statusCode: 500;
    @body error: Error;
  };

  @useAuth(BearerAuth)
  @tag("Landlord")
  @delete
  op archive(@path id: string): {
    @statusCode statusCode: 200;
    @body landlord: Landlord;
  } | {
    @statusCode statusCode: 401;
    @body error: Error;
  } | {
    @statusCode statusCode: 404;
    @body error: Error;
  } | {
    @statusCode statusCode: 500;
    @body error: Error;
  };
}

@route("/properties")
namespace Properties {
  @useAuth(BearerAuth)
  @tag("Property")
  @get
  op list(@query page?: int32, @query limit?: int32, @query address?: string, @query archived_only?: boolean): {
    @statusCode statusCode: 200;
    @body properties: PropertyList;
  } | {
    @statusCode statusCode: 500;
    @body error: Error;
  };

  @useAuth(BearerAuth)
  @tag("Property")
  @get
  op get(@path id: string): {
    @statusCode statusCode: 200;
    @body property: Property;
  } | {
    @statusCode statusCode: 401;
    @body error: Error;
  } | {
    @statusCode statusCode: 404;
    @body error: Error;
  } | {
    @statusCode statusCode: 500;
    @body error: Error;
  };

  @useAuth(BearerAuth)
  @tag("Property")
  @post
  op create(@body property: CreateProperty): {
    @statusCode statusCode: 201;
    @body property: Property;
  } | {
    @statusCode statusCode: 400;
    @body error: Error;
  } | {
    @statusCode statusCode: 500;
    @body error: Error;
  };

  @useAuth(BearerAuth)
  @tag("Property")
  @patch
  op update(@path id: string, @body property: UpdateProperty): {
    @statusCode statusCode: 200;
    @body property: Property;
  } | {
    @statusCode statusCode: 400;
    @body error: Error;
  } | {
    @statusCode statusCode: 401;
    @body error: Error;
  } | {
    @statusCode statusCode: 404;
    @body error: Error;
  } | {
    @statusCode statusCode: 500;
    @body error: Error;
  };

  @useAuth(BearerAuth)
  @tag("Property")
  @delete
  op archive(@path id: string): {
    @statusCode statusCode: 200;
    @body property: Property;
  } | {
    @statusCode statusCode: 401;
    @body error: Error;
  } | {
    @statusCode statusCode: 404;
    @body error: Error;
  } | {
    @statusCode statusCode: 500;
    @body error: Error;
  };
}

@route("/tenants")
namespace Tenants {
  @useAuth(BearerAuth)
  @tag("Tenant")
  @get
  op list(@query page?: int32, @query limit?: int32, @query name?: string, @query archived_only?: boolean): {
    @statusCode statusCode: 200;
    @body tenants: TenantList;
  } | {
    @statusCode statusCode: 500;
    @body error: Error;
  };

  @useAuth(BearerAuth)
  @tag("Tenant")
  @get
  op get(@path id: string): {
    @statusCode statusCode: 200;
    @body tenant: Tenant;
  } | {
    @statusCode statusCode: 401;
    @body error: Error;
  } | {
    @statusCode statusCode: 404;
    @body error: Error;
  } | {
    @statusCode statusCode: 500;
    @body error: Error;
  };

  @useAuth(BearerAuth)
  @tag("Tenant")
  @post
  op create(@body tenant: CreateTenant): {
    @statusCode statusCode: 201;
    @body tenant: Tenant;
  } | {
    @statusCode statusCode: 400;
    @body error: Error;
  } | {
    @statusCode statusCode: 500;
    @body error: Error;
  };

  @useAuth(BearerAuth)
  @tag("Tenant")
  @patch
  op update(@path id: string, @body tenant: UpdateTenant): {
    @statusCode statusCode: 200;
    @body tenant: Tenant;
  } | {
    @statusCode statusCode: 400;
    @body error: Error;
  } | {
    @statusCode statusCode: 401;
    @body error: Error;
  } | {
    @statusCode statusCode: 404;
    @body error: Error;
  } | {
    @statusCode statusCode: 500;
    @body error: Error;
  };

  @useAuth(BearerAuth)
  @tag("Tenant")
  @delete
  op archive(@path id: string): {
    @statusCode statusCode: 200;
    @body tenant: Tenant;
  } | {
    @statusCode statusCode: 401;
    @body error: Error;
  } | {
    @statusCode statusCode: 404;
    @body error: Error;
  } | {
    @statusCode statusCode: 500;
    @body error: Error;
  };
}